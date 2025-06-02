from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class Layout(BaseView):
    def __init__(self, *, children = None, header = None, footer = None):
        super().__init__()
        self.children = children
        self.header = header
        self.footer = footer

    def _render(self) -> Element:
        return el("div", [el("header", render_child(self.header) if self.header is not None else el("h1", "Default Header")), el("main", render_child(self.children) if self.children is not None else el("p", "Default content")), el("footer", render_child(self.footer) if self.footer is not None else el("p", "Default footer"))], {"class": "layout"})

class App(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _slot_children_1000 = []
        _slot_children_1000.append(el("h1", "Custom Header"))
        _slot_children_2000 = []
        _slot_children_2000.append(el("div", [el("p", "This is the main content"), el("p", "Multiple paragraphs")]))
        _slot_children_3000 = []
        _slot_children_3000.append(el("div", el("p", "&copy; 2024 My App")))
        return Layout(header=fragment(_slot_children_1000), children=fragment(_slot_children_2000), footer=fragment(_slot_children_3000))

