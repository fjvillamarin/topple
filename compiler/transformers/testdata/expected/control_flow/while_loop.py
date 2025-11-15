class CounterView(BaseView):
    def __init__(self, count):
        super().__init__()
        self.count = count

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        while self.count > 0:
            _div_children_2000.append(el("span", escape(self.count)))
            self.count = self.count - 1
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

