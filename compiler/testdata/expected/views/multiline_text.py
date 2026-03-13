from topple.psx import BaseView, Element, el, escape, fragment, raw
class SimpleMultiline(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("p", ["This is a longer paragraph", "that spans multiple lines."])

class MultilineWithInterpolation(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Welcome"))
        _div_children_2000.append(el("p", [f"Hello, {escape(self.name)}! Welcome to", "our website. We hope you enjoy", "your stay."]))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

class MultilineWithKeywords(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("p", ["This is not a drill and", "or maybe it is."])

class MultilineMixed(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("p", ["First paragraph with", "multiple lines."]))
        if True:
            _div_children_4000.append(el("span", "Conditional content"))
        _div_children_4000.append(el("p", ["Second paragraph with", "more text here."]))
        _root_children_3000.append(el("div", _div_children_4000))
        return fragment(_root_children_3000)

