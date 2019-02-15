// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"strings"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// Tag represents a group identifier which a Synse device can belong to.
type Tag struct {
	Namespace  string
	Annotation string
	Label      string

	string string
}

// NewTag creates a new Tag from a tag string.
func NewTag(tag string) *Tag {
	split := strings.SplitN(tag, "/", 2)
	namespace, component := split[0], split[1]

	split = strings.SplitN(component, ":", 2)
	annotation, label := split[0], split[1]

	return &Tag{
		Namespace:  namespace,
		Annotation: annotation,
		Label:      label,
		string:     tag,
	}
}

// String prints the Tag in its string representation.
func (tag *Tag) String() string {
	return tag.string
}

// Encode translates the Tag to its corresponding gRPC message.
func (tag *Tag) Encode() *synse.V3Tag {
	return &synse.V3Tag{
		Namespace:  tag.Namespace,
		Annotation: tag.Annotation,
		Label:      tag.Label,
	}
}
