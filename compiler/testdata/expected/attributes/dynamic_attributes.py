from topple.psx import BaseView, Element, el, escape, fragment
class DynamicAttributes(BaseView):
    def __init__(self, is_active: bool, user_id: int, css_class: str):
        super().__init__()
        self.is_active = is_active
        self.user_id = user_id
        self.css_class = css_class

    def _render(self) -> Element:
        _root_children_1000 = []
        item_count = 42
        _div_children_2000 = []
        _div_children_2000.append(el("input", "", {"type": "checkbox", "checked": escape(self.is_active)}))
        _div_children_2000.append(el("button", escape("Active" if self.is_active else "Inactive"), {"disabled": escape(not self.is_active)}))
        _div_children_2000.append(el("span", "Computed value", {"data-value": escape(self.user_id * 10)}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": escape(self.css_class), "data-user-id": escape(self.user_id), "data-count": escape(item_count)}))
        return fragment(_root_children_1000)

