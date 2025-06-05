from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class Panel(BaseView):
    def __init__(self, title: str, *, children = None, actions = None, footer = None):
        super().__init__()
        self.title = title
        self.children = children
        self.actions = actions
        self.footer = footer

    def _render(self) -> Element:
        _div_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h3", escape(self.title)))
        if has_slot("actions"):
            _div_children_2000.append(el("div", render_child(self.actions) if self.actions is not None else "", {"class": "panel-actions"}))
        _div_children_1000.append(el("div", _div_children_2000, {"class": "panel-header"}))
        _div_children_1000.append(el("div", render_child(self.children) if self.children is not None else "", {"class": "panel-body"}))
        if has_slot("footer"):
            _div_children_1000.append(el("div", render_child(self.footer) if self.footer is not None else "", {"class": "panel-footer"}))
        else:
            _div_children_1000.append(el("div", el("p", "No footer content provided."), {"class": "panel-footer text-muted"}))
        return el("div", _div_children_1000, {"class": "panel"})

class Example(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _slot_children_3000 = []
        _slot_children_3000.append(el("p", "This is the main content"))
        _slot_children_4000 = []
        _slot_children_4000.append(el("button", "Edit"))
        _slot_children_5000 = []
        _slot_children_5000.append(el("p", "Custom footer"))
        return Panel(title="My Panel", children=fragment(_slot_children_3000), actions=fragment(_slot_children_4000), footer=fragment(_slot_children_5000))

