// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// ActionTagKind is used to identify the Tag type
	ActionTagKind = "action"

	// ActionResultTagKind is used to identify the Tag type
	ActionResultTagKind = "actionresult"

	// actionMarker is the identifier used to join filterable
	// prefixes for Action Id's with unique suffixes
	actionMarker = "_a_"

	// actionResultMarker is the token used to delimit a filterable prefix from unique suffix
	actionResultMarker = "_ar_"
)

//
// ActionTag
//

// ActionTag is a Tag type for representing Action entities, which
// are records of queued actions for a given unit
type ActionTag struct {
	IdPrefixer
}

var _ PrefixTag = (*ActionTag)(nil)

// NewActionTag returns the tag for the action with the given id.
func NewActionTag(id string) ActionTag {
	tag, ok := newActionTag(id)
	if !ok {
		panic(fmt.Sprintf("%q is not a valid action id", id))
	}
	return tag
}

// JoinActionTag reconstitutes an ActionTag from it's prefix and sequence
func JoinActionTag(prefix string, sequence int) ActionTag {
	actionId := fmt.Sprintf("%s%s%d", prefix, actionMarker, sequence)
	tag, ok := newActionTag(actionId)
	if !ok {
		panic("bad prefix or sequence")
	}
	return tag
}

// IsValidAction returns whether actionId is a valid actionId
// Valid action ids include the names.actionMarker token that delimits
// a prefix that can be used for filtering, and a suffix that should be
// unique.  The prefix should match the name rules for units
func IsValidAction(actionId string) bool {
	return isValidIdPrefixTag(actionId, actionMarker)
}

// ParseActionTag parses a action tag string.
func ParseActionTag(actionTag string) (ActionTag, error) {
	tag, err := ParseTag(actionTag)
	if err != nil {
		return ActionTag{}, err
	}
	st, ok := tag.(ActionTag)
	if !ok {
		return ActionTag{}, invalidTagError(actionTag, ActionTagKind)
	}
	return st, nil
}

func newActionTag(actionId string) (ActionTag, bool) {
	if !isValidIdPrefixTag(actionId, actionMarker) {
		return ActionTag{}, false
	}
	prefixer := IdPrefixer{
		Id_:     actionId,
		Kind_:   ActionTagKind,
		Marker_: actionMarker,
	}
	return ActionTag{IdPrefixer: prefixer}, true
}

//
// ActionResultTag
//

// ActionResultTag represents the actionresult of an action
type ActionResultTag struct {
	IdPrefixer
}

var _ PrefixTag = (*ActionResultTag)(nil)

// NewActionResultTag returns a tag for an actionresult using it's id
func NewActionResultTag(id string) ActionResultTag {
	tag, ok := newActionResultTag(id)
	if !ok {
		panic(fmt.Sprintf("%q is not a valid action result id", id))
	}
	return tag
}

// IsValidActionResult returns whether resultId is a valid actionResultId
// Valid action result ids include the names.actionResultMarker token that delimits
// a prefix that can be used for filtering, and a suffix that should be
// unique. The prefix should match the name rules for units or services
func IsValidActionResult(resultId string) bool {
	return isValidIdPrefixTag(resultId, actionResultMarker)
}

// ParseActionResultTag parses a action result tag string.
func ParseActionResultTag(actionResultTag string) (ActionResultTag, error) {
	tag, err := ParseTag(actionResultTag)
	if err != nil {
		return ActionResultTag{}, err
	}
	st, ok := tag.(ActionResultTag)
	if !ok {
		return ActionResultTag{}, invalidTagError(actionResultTag, ActionResultTagKind)
	}
	return st, nil
}

func newActionResultTag(resultId string) (ActionResultTag, bool) {
	if !isValidIdPrefixTag(resultId, actionResultMarker) {
		return ActionResultTag{}, false
	}
	prefixer := IdPrefixer{
		Id_:     resultId,
		Kind_:   ActionResultTagKind,
		Marker_: actionResultMarker,
	}
	return ActionResultTag{IdPrefixer: prefixer}, true
}

//
// IdPrefixer
//

type PrefixTag interface {
	Tag
	Prefix() string
	Sequence() int
	PrefixTag() Tag
}

// IdPrefixer is a common type for representing tags that have
// structured prefixes.
// Note: this type is only used as an embedded type in other Tags that
// use structured prefixes, like ActionTag, and ActionResultTag. We
// have to export this type however so that gccgo is able to serialize
// the embedding types without dropping these fields.
// see: https://bugs.launchpad.net/juju-core/+bug/1381626
type IdPrefixer struct {
	Id_     string
	Kind_   string
	Marker_ string
}

var _ PrefixTag = (*IdPrefixer)(nil)

// Id returns the id of the type this Tag represents
func (t IdPrefixer) Id() string { return t.Id_ }

// String returns a string that shows the type and id of the Tag
func (t IdPrefixer) String() string { return t.Kind_ + "-" + t.Id() }

// Kind exposes the value to identify what kind of Tag this is
func (t IdPrefixer) Kind() string { return t.Kind_ }

// Prefix returns the string representation of the prefix of the Tag
func (t IdPrefixer) Prefix() string {
	prefix, _, ok := splitId(t.Id(), t.Marker_)
	if !ok {
		return ""
	}
	return prefix
}

// Sequence returns the unique integer suffix of the Tag
func (t IdPrefixer) Sequence() int {
	_, sequence, ok := splitId(t.Id(), t.Marker_)
	if !ok {
		return -1
	}
	return sequence
}

// PrefixTag returns a Tag representing the Entity matching the id
// prefix
func (t IdPrefixer) PrefixTag() Tag {
	prefix, _, ok := splitId(t.Id(), t.Marker_)
	if !ok {
		return nil
	}

	var tag Tag
	var err error

	switch {
	case IsValidUnit(prefix):
		tag = NewUnitTag(prefix)
	case IsValidService(prefix):
		tag = NewServiceTag(prefix)
	default:
		tag, err = ParseTag(prefix)
		if err != nil {
			tag = nil
		}
	}
	return tag
}

// isValidIdPrefixTag signals whether the id is a validly formatted id
// for an IdPrefixer with the given marker
func isValidIdPrefixTag(id, marker string) bool {
	prefix, _, ok := splitId(id, marker)
	if !ok {
		return false
	}
	switch {
	case IsValidUnit(prefix):
	case IsValidService(prefix):
	default:
		return false
	}
	return true
}

// splitId extracts the prefix and suffix from the id using the marker
// token
func splitId(id, marker string) (string, int, bool) {
	parts := strings.Split(id, marker)
	if len(parts) != 2 {
		return "", 0, false
	}
	if len(parts[1]) > 1 && parts[1][:1] == "0" {
		return "", 0, false
	}
	seq, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return "", 0, false
	}
	return parts[0], int(seq), true
}
