from psx.runtime import BaseView, Element, el, escape, fragment
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
        _root_children_1000.append(el("div", [el("h2", f"Items ({escape(len(self.items))})"), ""], {"class": "items-list"}))
        return fragment(_root_children_1000)

