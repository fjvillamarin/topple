from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class DynamicAttributes(BaseView):
    def __init__(self, is_active: bool, user_id: int, css_class: str):
        super().__init__()
        self.is_active = is_active
        self.user_id = user_id
        self.css_class = css_class

    def _render(self) -> Element:
        _view_children_1000 = []
        item_count = 42
        _view_children_1000.append(el("div", [el("input", "", {"type": "checkbox", "checked": escape(self.is_active)}), el("button", "", {"disabled": escape(not self.is_active)}), el("span", "Computed value", {"data-value": escape(self.user_id * 10)})], {"class": escape(self.css_class), "data-user-id": escape(self.user_id), "data-count": escape(item_count)}))
        return fragment(_view_children_1000)

