#!/usr/bin/env python3
import sys
sys.path.insert(0, '.')

from topple.psx import BaseView, Element, el, escape, fragment, raw

class ExampleCard(BaseView):
    def __init__(self, label: str, code: str):
        super().__init__()
        self.label = label
        self.code = code

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("h4", escape(self.label), {"class": "text-purple-400 font-mono text-sm mb-3"}))
        _div_children_4000.append(el("pre", el("code", escape(raw(self.code))), {"class": "text-slate-300 text-xs font-mono whitespace-pre overflow-x-auto"}))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "bg-slate-900 rounded-lg p-4"}))
        return fragment(_root_children_3000)

# Test data
EXAMPLES = [
    {
        "label": "For Loops",
        "code": "for item in items:\n    <li>{item}</li>"
    },
    {
        "label": "Conditionals",
        "code": "if show:\n    <div>Visible</div>"
    },
    {
        "label": "Nested Elements",
        "code": "for row in rows:\n    <tr>\n        for cell in row:\n            <td>{cell}</td>\n    </tr>"
    }
]

import html

print("=== Testing ExampleCard with raw() ===\n")
for example in EXAMPLES:
    card = ExampleCard(label=example["label"], code=example["code"])
    output = card.render()

    # Extract the code content
    code_content = output.split('<code>')[1].split('</code>')[0]
    decoded = html.unescape(code_content)

    print(f"✓ {example['label']}:")
    print(decoded)
    print()
