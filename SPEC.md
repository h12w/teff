Specification of Test Friendly Format
=====================================

Hǎiliàng Wáng <w@h12.me>

Copyright (c) 2014-2015, Hǎiliàng Wáng. All rights reserved.

This specification is licensed under the Creative Commons Attribution 4.0
International License. To view a copy of this license, visit
    http://creativecommons.org/licenses/by/4.0/

Introduction
------------
TEFF (TEst Friendly Format) is an extensible data format with testing purpose in
mind. it is easy to read, compare and write manually.

In general, the [core format](#core) of TEFF represents a tree. Each node of the
tree is a string occupying a single line, and the relation between nodes are
represented by indents.

This model is simple and [extensible](#extensions). The minimal constraints make
it possible to extend the resprentation of a data structure without intefering
other nodes.

This specification is a followup work of [OGDL 2.0](https://github.com/ogdl)
(OGDL was invented by Rolf Veen, and we cooperated in writing its 2.0 spec).
The major difference between TEFF and OGDL is that TEFF disallow mutiple values
occupying a single line. This constraint simplifies the parser, opens more
possibilities for extention and makes it easier to compare two files line by line.

Notation
--------
The syntax is specified using a variant of Extended Backus-Naur Form, based on
[W3C XML EBNF](http://www.w3.org/TR/xml11/#sec-notation),
which is extended with the following definitions:
* `EOF` matches the end of the file.
* `LINE_START` matches the start of a line.
* **Escape sequences** defined in section [Interpreted string](#interpreted-string).
* Regular expressions defined by [Golang Regexp](http://golang.org/pkg/regexp/syntax/).
* Text enclosed by <> is a description.

Core
----
A TEFF file is a sequence of [Unicode](http://unicode.org/) code points encoded
in [UTF-8](http://www.unicode.org/versions/latest/ch03.pdf).

Except `\t` (U+0009), `\n` (U+000A) and `\r` (U+000D), code points less than
U+0032 are invalid and should not appear in a TEFF file.

    char_visible   ::= [^\x00-\x20]
    char_space     ::= [ \t]
    char_inline    ::= char_visible | char_space
    char_break     ::= [\r\n]
    char_any       ::= char_inline | char_break
    spaces         ::= char_space+
    lead_space     ::= LINE_START spaces
    unicode_letter ::= <a Unicode code point classified as "Letter">
    unicode_digit  ::= <a Unicode code point classified as "Decimal Digit">
    letter_digit   ::= unicode_letter | unicode_digit | "_"

TEFF tokens:

    spaces       ::= char_space+
    annotation   ::= lead_space? "#" char_inline* (newline | EOF)
    newline      ::= char_break | "\r\n"
    string       ::= <one or more consecutive char_inline's excluding the lead_space>
    indent       ::= <an indent token is emitted when the length of the lead_space
                      increases in this line compared to the previous line>
    unindent     ::= <one or more unindent tokens are emitted when the length of
                      the lead_space decreases in this line compared to the previous
                      line. Each of them cancels the last indent, till the indent
                      level becomes the same as the next line>

TEFF grammer:

    teff_file    ::= list EOF
    list         ::= node*
    node         ::= value (indent list unindent)?
    value        ::= string

Note: TEFF grammar only cares about tokens of type `string`, `indent` and `unindent`.
Accumulated `annocation` tokens should be attached to the next node.

Extensions
----------
In this section, format extensions for common types are specified. These types
should cover all the builtin types and some of the types in standard libraries.

### Reference & type annotation
TEFF can represent a cyclic graph by referenced annotation.

    ref_id          ::= "^" letter_digit+
    ref_annotation  ::= lead_space? "#" spaces? ref_id neweline
     ↓                   ↓              ↓             ↓
    annotation      ::= lead_space? "#" char_inline*  (newline | EOF)

`ref_id` is a unique ID within a TEFF file. It should be defined only once but
can be referenced multiple times by the `ref_id`.

TEFF can optionally represent type by using type annotation.

    type_label      ::= "<" letter_digit+ ">"
    type_annotation ::= lead_space? "#" spaces? type_label newline
     ↓                   ↓              ↓                 ↓
    annotation      ::= lead_space? "#" char_inline*      (newline | EOF)

When both a cyclic reference and a type are defined for a node, it does not
matter which comes first. Both annotates the next node.

### Array

An array is represented as a list.

    array         ::= array_element*
     ↓                 ↓
    list          ::= node*

To represent an array of array, the anonymous symbol `_` is introduced to
represent the anonymous parent of a child array.

    array_element ::=  _     indent array unindent
     ↓                 ↓             ↓
    node          ::= value (indent list  unindent)?

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

    map           ::= key_value*
     ↓                 ↓
    list          ::= node*

    key_value     ::= map_key indent map_value unindent
     ↓                 ↓              ↓
    node          ::= value  (indent list      unindent)?

Some languages (like Go) support a compound map key like array of struct. It can
be represented in TEFF as long as the key can be encoded into a single line string.
The encoding is implementation specific, and will be treated as a normal string
for languages that do not support compound map key.

### Nil

The special string nil is used to represent an uninitialized nullable node.

    nil    ::= "nil"
     ↓          ↓
    value  ::= string

### Interpreted string
Interpreted string is a double quoted string, that can interpret certain escape
sequences.

    quoted_char        ::= (char_inline - '"') | '\\"'
    interpreted_string ::= '"' (unicode_value | byte_value)* '"'
     ↓                      ↓
    value              ::= string

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
     ↓              ↓
    value      ::= string

### Numeric value
Numeric value is an unquoted string that encode a number.

    sign       ::= "+" | "-"
    decimals   ::= [1-9] [0-9]*

#### Integer
    integer    ::= sign? decimals
     ↓              ↓
    value      ::= string

#### Float
Float value is an unquoted string that encode a floating point number:

    exponent   ::= ( "e" | "E" ) ( "+" | "-" )? decimals
    float_base ::= (decimals "." decimal* exponent?) |
                   (decimals exponent) |
                   ("." decimals exponent?)
    float      ::= sign? float_base
     ↓              ↓
    value      ::= string

#### Complex
    int_float  ::= decimals | float_base
    complex    ::= sign? int_float sign int_float "i"
     ↓              ↓
    value      ::= string

### Date/time (TODO: use a shorter representation)
A date/time value is an unquoted string encoded with
[RFC3339](http://www.rfc-editor.org/rfc/rfc3339.txt)

    date_time  ::= rfc3339_date_time
     ↓              ↓
    value      ::= string

e.g.

    2006-01-02T15:04:05.999999999Z07:00

### IP address
An IP address is either an IPv4 or IPv6 address.

    ip         ::= ipv4 | ipv6

An IPv4 address value is an unquoted string encoded with dot-decimal notation:

    ipv4       ::= decimals "." decimals "." decimals "." decimals
     ↓              ↓
    value      ::= string

e.g.

    74.125.19.99

An IPv6 address value is an unquoted string encoded with
[RFC5952](http://www.rfc-editor.org/rfc/rfc5952.txt).

    ipv6       ::= rfc5952_ipv6_address
     ↓              ↓
    value      ::= string

e.g.

    2001:4860:0:2001::68

