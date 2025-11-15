from topple.psx import BaseView, Element, el, escape, fragment
class SafeDisplay(BaseView):
    def __init__(self, value: str):
        super().__init__()
        self.value = value

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        try:
            _div_children_2000.append(el("p", f"Value: {escape(int(self.value))}"))
        except ValueError:
            _div_children_2000.append(el("p", "Invalid value"))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

