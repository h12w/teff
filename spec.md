Specification of Test Friendly Format
=====================================

Hǎiliàng Wáng <w@h12.me>

Copyright (c) 2014, Hǎiliàng Wáng. All rights reserved.

This specification is licensed under the Creative Commons Attribution 4.0
International License. To view a copy of this license, visit
    http://creativecommons.org/licenses/by/4.0/

Introduction
------------
TFF (Test Friendly Format) is an extensible data format with testing purpose in
mind. it is easy to read, compare and write manually.

In general, the format of represents a tree. Each node of the tree is a string
occupying a single line, and the relation between nodes are represented by indents.

This model is simple and extensible. The minimal constraints make it possible to
extend the resprentation of a data structure without intefering other nodes.

This specification is a followup work of [OGDL 2.0](https://github.com/ogdl)
(OGDL was invented by Rolf Veen, and we cooperated in writing its 2.0 spec).
The major difference between TFF and OGDL is that TFF disallow mutiple
values occupying a single line. This constraint simplifies the parser, opens more
possibilities for extention and makes it easier to compare two files line by line.

Notation
--------
The syntax is specified using a variant of Extended Backus-Naur Form, based on
[W3C XML EBNF](http://www.w3.org/TR/2006/REC-xml11-20060816/#sec-notation),
which is extended with the following definitions:
* **EOF** matches the end of the file.
* **Escape sequences** defined in section [Interpreted string](#interpreted-string).
* Regular expressions defined by [Golang Regexp](http://golang.org/pkg/regexp/syntax/).
* Text enclosed by <> is a description.

Core
----
A TFF file is a sequence of [Unicode](http://unicode.org/) code points encoded
in [UTF-8](http://www.unicode.org/versions/Unicode6.2.0/ch03.pdf).

Except \t (U+0009), \n (U+000A) and \r (U+000D), code points less than U+0032 are
invalid and should not appear in a TFF file.

    char_visible ::= [^\x00-\x20]
    char_space   ::= [ \t]
    char_inline  ::= char_visible | char_space
    char_break   ::= [\r\n]
    char_any     ::= char_inline | char_break
    lead_space   ::= <one or more char_space's at the start of a line>

TFF tokens:

    comment      ::= "#" char_inline* (newline | EOF)
    newline      ::= char_break | "\r\n"
    string       ::= <one or more consecutive char_inline's excluding the lead_space>
    indent       ::= <an indent token is emitted when the length of the lead_space increases in this line compared to the previous line>
    unindent     ::= <one or more unindent tokens are emitted when the length of the lead_space decreases in this line compared to the previous line. Each of them cancels the last indent, till the indent level becomes the same as the next line>

A TFF parser only cares about tokens of type 'string', 'indent' and 'unindent'.

    tff_file     ::= list EOF
    list         ::= node*
    node         ::= value (indent list unindent)?
    value        ::= string

Extensions
----------
Core definitions only include tree nodes, however, richer data structures can be
mapped onto the two primitives.

In this section, format for common types are specified. These types are intended
to be mapped to builtin types or types in standard libraries, and format of
these types are fixed and cannot be overriden.

### Array

An array is represented as a list.

    list          ::= node*
     ↓                 ↓
    array         ::= array_element*

To represent an array of array, the anonymous symbol "_" is introduced.

    node          ::= value (indent list unindent)?
     ↓                 ↓             ↓
    array_element ::=  _    indent array unindent

e.g.

    -
        1
        2
        3
    -
        4
        5

### Map
A map is represented with a list of key-value pairs. Each pair is represented
as a parent-child relation.

    list          ::= node*
     ↓                 ↓
    map           ::= key_value*

    node          ::= value  (indent list unindent)?
     ↓                 ↓              ↓
    key_value     ::= map_key indent map_value unindent

Some languages (like Go) support a compound map key like array of struct. It can
be represented in TFF as long as the key can be encoded into a single line string.
The encoding is implementation specific, and will be treated as a normal string
for languages that do not support compound map key.

### Referenced & typed node
TFF can represent a cyclic graph by referenced nodes.

    ref_id          ::= '^' char_visible+
    referenced_node ::= ref_id (indent node* unindent)

A cyclic reference id is a unique ID within a TFF file. It should be defined
only once but can be referenced multiple times by the reference ID alone.

TFF can (optionally) represent type by using typed nodes.

    type            ::= '!' char_visible+
    typed_node      ::= type (indent node* unindent)

When both a cyclic reference and a type are defined for a node, it does not
matter which comes first.

### Nil

The special string nil is used to represent an uninitialized nullable node.

    nil ::= "nil"

### Interpreted string
Interpreted string is a double quoted string, that can interpret certain escape
sequences.

    quoted_char        ::= (char_inline - '"') | '\\"'
    interpreted_string ::= '"' quoted_char* '"'

Escape sequences:

    \a    U+0007 alert or bell
    \b    U+0008 backspace
    \t    U+0009 horizontal tab
    \n    U+000A line feed or newline
    \v    U+000B vertical tab
    \f    U+000C form feed
    \r    U+000D carriage return
    \\    U+005C backslash
    \"    U+0022 double quote "
    \x    Unicode code point represented with two hexadecimal digits followed by \x
    \u    Unicode code point represented with exactly 4 hexadecimal digits followed by \u
    \U    Unicode code point represented with exactly 8 hexadecimal digits followed by \U

### Boolean value
Boolean value is an unquoted string of either true of false.

    boolean    ::= "true" | "false"

### Numeric value
Numeric value is an unquoted string that encode a number.

    sign       ::= "+" | "-"
    decimals   ::= [1-9] [0-9]*

#### Integer
    integer    ::= sign? decimals

#### Float
Float value is an unquoted string that encode a floating point number:

    exponent   ::= ( "e" | "E" ) ( "+" | "-" )? decimals
    float_base ::= (decimals "." decimal* exponent?) |
                   (decimals exponent) |
                   ("." decimals exponent?)
    float      ::= sign? float_base

#### Complex
    int_float  ::= decimals | float_base
    complex    ::= sign? int_float sign int_float "i"

### Date/time
A date/time value is an unquoted string encoded with
[RFC3339](http://www.rfc-editor.org/rfc/rfc3339.txt)

    date_time  ::= rfc3339_date_time

e.g.

    2006-01-02T15:04:05.999999999Z07:00

### IP address
An IP address is either an IPv4 or IPv6 address.

    ip         ::= ipv4 | ipv6

An IPv4 address value is an unquoted string encoded with dot-decimal notation:

    ipv4       ::= decimals "." decimals "." decimals "." decimals

e.g.

    74.125.19.99

An IPv6 address value is an unquoted string encoded with
[RFC5952](http://www.rfc-editor.org/rfc/rfc5952.txt).

    ipv6       ::= rfc5952_ipv6_address

e.g.

    2001:4860:0:2001::68

