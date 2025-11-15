from topple.psx import BaseView, Element, el, escape, fragment
class Counter(BaseView):
    def __init__(self, start: int, end: int):
        super().__init__()
        self.start = start
        self.end = end

    def _render(self) -> Element:
        _root_children_1000 = []
        _ul_children_2000 = []
        i = self.start
        while i < self.end:
            _ul_children_2000.append(el("li", escape(i)))
            i = i + 1
        _root_children_1000.append(el("ul", _ul_children_2000))
        return fragment(_root_children_1000)

