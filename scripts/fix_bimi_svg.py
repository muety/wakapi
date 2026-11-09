#!/usr/bin/env python3
"""
Convert an Inkscape/SVG 1.1 export into an SVG 1.2 Portable Subset document.
Fixes validation errors to conform to what BIMI records require (https://bimigroup.org/using-the-rnc-schema-to-validate-bimi-svg-images/).
Validate via:
    - wget http://bimigroup.org/resources/SVG_PS-latest.rnc.txt
    - uv tool install jingtrang
    - uvx --from jingtrang pyjing -c SVG_PS-latest.rnc.txt icon.svg

Usage: to_svg_ps.py IN.svg [-o OUT.svg] [--title "Name"]

Note: this script is nearly 99 % AI generated.
"""

import argparse
import re
import sys
import xml.etree.ElementTree as ET

SVG = "http://www.w3.org/2000/svg"
ET.register_namespace("", SVG)

PS_VERSION = "1.2"
PS_BASE_PROFILE = "tiny-ps"

CSS_TO_PRESENTATION = re.compile(r"^[a-z-]+$")


def q(tag: str) -> str:
    return f"{{{SVG}}}{tag}"


def style_to_attributes(el: ET.Element) -> None:
    style = el.attrib.pop("style", None)
    if not style:
        return
    for rule in style.split(";"):
        if ":" not in rule:
            continue
        name, _, value = rule.partition(":")
        name, value = name.strip(), value.strip()
        if not CSS_TO_PRESENTATION.match(name):
            # @font-face style selectors, animations etc. have no attribute form
            print(f"warning: dropped unsupported style declaration {name!r}", file=sys.stderr)
            continue
        el.set(name, value)


def strip_clipping(root: ET.Element) -> None:
    for el in root.iter():
        el.attrib.pop("clip-path", None)
        el.attrib.pop("clip-rule", None)
        # clipPathUnits only exists on clipPath itself, nothing to do here
    for parent in list(root.iter()):
        for child in list(parent):
            if child.tag in (q("clipPath"), q("mask"), q("filter")):
                parent.remove(child)


def drop_empty(root: ET.Element) -> None:
    """Remove structural elements that no longer render anything."""
    removable = {q("defs"), q("g")}
    changed = True
    while changed:
        changed = False
        for parent in list(root.iter()):
            for child in list(parent):
                if child.tag in removable and len(child) == 0:
                    parent.remove(child)
                    changed = True


def convert(src: str, dst: str, title: str) -> None:
    tree = ET.parse(src)
    root = tree.getroot()

    # <metadata> in the portable subset may contain text only, so the RDF block
    # cannot survive; it carries no licence grant in Inkscape exports, just
    # dc:format / dc:type boilerplate.
    for child in list(root):
        if child.tag == q("metadata"):
            root.remove(child)

    strip_clipping(root)

    for el in root.iter():
        style_to_attributes(el)

    drop_empty(root)

    # required attributes and the mandatory first child
    root.set("version", PS_VERSION)
    root.set("baseProfile", PS_BASE_PROFILE)
    for child in list(root):
        if child.tag == q("title"):
            root.remove(child)
    title_el = ET.Element(q("title"))
    title_el.text = title
    root.insert(0, title_el)

    # namespace declarations that are no longer referenced: ElementTree keeps
    # literal xmlns:foo="..." attributes under the xmlns axis
    for name in list(root.attrib):
        if name.startswith("{http://www.w3.org/2000/xmlns/}"):
            prefix = name.rsplit("}", 1)[1]
            if prefix in ("sodipodi", "inkscape", "rdf", "cc", "dc", "svg"):
                del root.attrib[name]

    ET.indent(tree, space="  ")
    tree.write(dst, encoding="UTF-8", xml_declaration=True)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("svg")
    ap.add_argument("-o", "--out", default=None)
    ap.add_argument("--title", default="Icon")
    args = ap.parse_args()
    out = args.out or re.sub(r"\.svg$", "", args.svg) + ".ps.svg"
    convert(args.svg, out, args.title)
    print(f"wrote {out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
