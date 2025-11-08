from topple.psx import BaseView, Element, el, escape, fragment
class EarlyReturnView(BaseView):
    def __init__(self, items: list, show_empty: bool=True):
        super().__init__()
        self.items = items
        self.show_empty = show_empty

    def _render(self) -> Element:
        _root_children_1000 = []
        if not self.items and not self.show_empty:
            _root_children_1000.append(el("div", "No content to display", {"class": "empty-message"}))
            return
        if not self.items:
            _root_children_1000.append(el("div", [el("h2", "No items found"), el("p", "Try adding some items")], {"class": "empty-state"}))
            return
        _div_children_2000 = []
        _div_children_2000.append(el("h2", f"Items ({escape(len(self.items))})"))
        for item in self.items:
            _div_children_2000.append(el("div", escape(item), {"class": "item"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "items-list"}))
        return fragment(_root_children_1000)

