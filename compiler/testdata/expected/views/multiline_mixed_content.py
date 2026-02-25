from topple.psx import BaseView, Element, el, escape, fragment, raw
class MixedContent(BaseView):
    def __init__(self, items: list, show_header: bool):
        super().__init__()
        self.items = items
        self.show_header = show_header

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("p", f"Here is some text with {escape(len(self.items))} items below."))
        if self.show_header:
            _div_children_2000.append(el("h1", "Items List"))
        for item in self.items:
            _div_children_2000.append(el("li", escape(item)))
        _div_children_2000.append(el("footer", f"End of list with {escape(len(self.items))} total items."))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

