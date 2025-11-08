class CounterView(BaseView):
    def __init__(self, count):
        super().__init__()
        self.count = count

    def _render(self) -> Element:
        _root_children_2000 = []
        _div_children_3000 = []
        while count > 0:
            _div_children_3000.append(el("span", ""))
            count = count - 1
        _root_children_2000.append(el("div", _div_children_3000))
        return fragment(_root_children_2000)

