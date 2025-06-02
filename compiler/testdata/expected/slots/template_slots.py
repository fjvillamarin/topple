from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class Card(BaseView):
    def __init__(self, *, children = None, header = None, footer = None):
        super().__init__()
        self.children = children
        self.header = header
        self.footer = footer

    def _render(self) -> Element:
        return el("div", [el("div", render_child(self.header) if self.header is not None else el("h4", "Default Header"), {"class": "card-header"}), el("div", render_child(self.children) if self.children is not None else "", {"class": "card-body"}), el("div", render_child(self.footer) if self.footer is not None else "", {"class": "card-footer"})], {"class": "card"})

class App(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _slot_children_1000 = []
        _slot_children_1000.append(el("template", [el("h2", "Complex Header"), el("p", "With subtitle")]))
        _slot_children_2000 = []
        _slot_children_2000.append(el("p", "Main content"))
        _slot_children_2000.append(el("p", "More content"))
        _slot_children_3000 = []
        _slot_children_3000.append(el("template", [el("button", "Save"), el("button", "Cancel")]))
        return Card(header=fragment(_slot_children_1000), children=fragment(_slot_children_2000), footer=fragment(_slot_children_3000))

