// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	gitolitev1alpha1 "heimstatt.dev/apis/gitolite/v1alpha1"
)

// Reasons a definition is left out of the rendered configuration.
var (
	// ErrInvalidName is a repository, group, user, option or config name
	// gitolite would not accept.
	ErrInvalidName = errors.New("invalid name")

	// ErrInvalidRef is a refex gitolite would not accept.
	ErrInvalidRef = errors.New("invalid ref")

	// ErrInvalidPermission is a permission gitolite would not accept.
	ErrInvalidPermission = errors.New("invalid permission")

	// ErrInvalidValue is an option, config or description value gitolite
	// would not accept.
	ErrInvalidValue = errors.New("invalid value")

	// ErrDisallowedConfigKey is a git config name GIT_CONFIG_KEYS does not
	// allow.
	ErrDisallowedConfigKey = errors.New("git config key not allowed")
)

// hashLength is the number of hex digits kept of the SHA-256, 64 bits.
const hashLength = 16

// descriptionKey is the git config behind gitolite's "desc".
const descriptionKey = "gitweb.description"

// optionKeyPrefix is the git config behind gitolite's "option".
const optionKeyPrefix = "gitolite-options."

// optionsKeyPattern is allowed by gitolite whatever GIT_CONFIG_KEYS says.
const optionsKeyPattern = `gitolite-options\..*`

// Rendered is the outcome of rendering a Config.
type Rendered struct {
	// Conf is the content of gitolite.conf.
	Conf string

	// RC is the content of .gitolite.rc.
	RC string

	// Hash identifies Conf and RC together.
	Hash string

	// RejectedGroups are the groups left out of Conf, by name.
	RejectedGroups map[string]error

	// RejectedRepositories are the repositories left out of Conf, by name.
	RejectedRepositories map[string]error
}

// Render turns config into the files gitolite compiles. Definitions gitolite
// would refuse are left out and reported in the result; the error is for a
// Config nothing can be rendered from.
func Render(config Config) (Rendered, error) {
	allowedKeys, err := compileKeys(config.Settings.GitConfigKeys)
	if err != nil {
		return Rendered{}, err
	}

	rendered := Rendered{
		RejectedGroups:       map[string]error{},
		RejectedRepositories: map[string]error{},
	}

	var blocks []string
	if groups := renderGroups(config.Groups, rendered.RejectedGroups); len(groups) > 0 {
		blocks = append(blocks, strings.Join(groups, ""))
	}
	for _, repository := range config.Repositories {
		block, err := renderRepository(repository, allowedKeys)
		if err != nil {
			rendered.RejectedRepositories[repository.Name] = err
			continue
		}
		blocks = append(blocks, block)
	}

	rendered.Conf = strings.Join(blocks, "\n")
	rendered.Hash = hash(rendered.Conf, rendered.RC)
	return rendered, nil
}

// compileKeys compiles the allowed git config names the way gitolite matches
// them: anchored and case-insensitive.
func compileKeys(patterns []string) ([]*regexp.Regexp, error) {
	patterns = append(slices.Clone(patterns), optionsKeyPattern)
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		expression, err := regexp.Compile("(?i)^(?:" + pattern + ")$")
		if err != nil {
			return nil, fmt.Errorf("compiling git config key pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, expression)
	}
	return compiled, nil
}

func isAllowedKey(key string, allowed []*regexp.Regexp) bool {
	for _, expression := range allowed {
		if expression.MatchString(key) {
			return true
		}
	}
	return false
}

func hash(conf, rc string) string {
	sum := sha256.Sum256([]byte(conf + "\x00" + rc))
	return hex.EncodeToString(sum[:])[:hashLength]
}

// renderGroups renders the groups that have members:
// one without cannot be written in gitolite's syntax.
func renderGroups(groups []Group, rejected map[string]error) []string {
	var lines []string
	for _, group := range groups {
		if err := checkGroup(group); err != nil {
			rejected[group.Name] = err
			continue
		}
		if len(group.Members) == 0 {
			continue
		}
		lines = append(lines, groupSigil+group.Name+" = "+strings.Join(group.Members, " ")+"\n")
	}
	return lines
}

func checkGroup(group Group) error {
	if group.Name == allGroup {
		return fmt.Errorf("%w: group %q is built in", ErrInvalidName, group.Name)
	}
	if err := checkCharacters(group.Name, "", nameCharacters); err != nil {
		return fmt.Errorf("%w: group %q %w", ErrInvalidName, group.Name, err)
	}
	for _, member := range group.Members {
		if strings.HasPrefix(member, groupSigil) {
			return fmt.Errorf("%w: member %q is a group", ErrInvalidName, member)
		}
		if err := checkCharacters(member, patternStart, nameCharacters+patternCharacters); err != nil {
			return fmt.Errorf("%w: member %q %w", ErrInvalidName, member, err)
		}
	}
	return nil
}

func renderRepository(repository Repository, allowedKeys []*regexp.Regexp) (string, error) {
	if err := validateRepository(repository.Name); err != nil {
		return "", err
	}

	var block strings.Builder
	block.WriteString("repo " + repository.Name + "\n")

	for i, rule := range repository.Rules {
		line, err := renderRule(rule)
		if err != nil {
			return "", fmt.Errorf("rule %d: %w", i, err)
		}
		block.WriteString("    " + line + "\n")
	}

	// Options and the description are gitolite's sugar for git config.
	config := map[string]string{}
	maps.Copy(config, repository.GitConfig)
	for key, value := range repository.Options {
		config[optionKeyPrefix+key] = value
	}
	if strings.TrimSpace(repository.Description) != "" {
		config[descriptionKey] = repository.Description
	}

	for _, key := range slices.Sorted(maps.Keys(config)) {
		if err := checkCharacters(key, "", keyCharacters); err != nil {
			return "", fmt.Errorf("%w: git config %q %w", ErrInvalidName, key, err)
		}
		if !isAllowedKey(key, allowedKeys) {
			return "", fmt.Errorf("%w: %q", ErrDisallowedConfigKey, key)
		}
		if err := validateValue(config[key]); err != nil {
			return "", fmt.Errorf("git config %q: %w", key, err)
		}
		// An empty value unsets the key; gitolite trims the line anyway.
		block.WriteString(strings.TrimRight("    config "+key+" = "+config[key], " ") + "\n")
	}

	return block.String(), nil
}

func renderRule(rule gitolitev1alpha1.RepositoryRule) (string, error) {
	if err := validatePermission(rule.Permission); err != nil {
		return "", err
	}
	permission := rule.Permission
	if rule.Denied {
		permission = "-"
	}

	for _, ref := range rule.Refs {
		if err := validateRef(ref); err != nil {
			return "", err
		}
	}

	// gitolite ignores a rule without subjects, which would drop it silently.
	if len(rule.Subjects) == 0 {
		return "", fmt.Errorf("%w: no subjects", ErrInvalidName)
	}
	for _, subject := range rule.Subjects {
		name := strings.TrimPrefix(subject, groupSigil)
		if err := checkCharacters(name, "", nameCharacters); err != nil {
			return "", fmt.Errorf("%w: subject %q %w", ErrInvalidName, subject, err)
		}
	}

	fields := append([]string{permission}, rule.Refs...)
	fields = append(fields, "=")
	fields = append(fields, rule.Subjects...)
	return strings.Join(fields, " "), nil
}
