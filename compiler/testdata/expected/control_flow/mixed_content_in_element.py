from topple.psx import BaseView, Element, el, escape, fragment, raw
class TodoList(BaseView):
    def __init__(self, items: list):
        super().__init__()
        self.items = items

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Todo List"))
        for item in self.items:
            _div_children_2000.append(el("li", escape(item)))
        _div_children_2000.append(el("footer", "End of list"))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

