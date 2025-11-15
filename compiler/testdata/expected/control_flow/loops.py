from topple.psx import BaseView, Element, el, escape, fragment
class LoopView(BaseView):
    def __init__(self, items: list, max_count: int=10):
        super().__init__()
        self.items = items
        self.max_count = max_count

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h2", "Items List"))
        _ul_children_3000 = []
        for (i, item) in enumerate(self.items):
            if i >= self.max_count:
                break
            _li_children_4000 = []
            _li_children_4000.append(el("span", f"{escape(i + 1)}. {escape(item)}"))
            if i % 2 == 0:
                _li_children_4000.append(el("span", "Even", {"class": "even"}))
            _ul_children_3000.append(el("li", _li_children_4000))
        _div_children_2000.append(el("ul", _ul_children_3000))
        counter = 0
        while counter < 3:
            _div_children_2000.append(el("p", f"Counter: {escape(counter)}"))
            counter = counter + 1
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

