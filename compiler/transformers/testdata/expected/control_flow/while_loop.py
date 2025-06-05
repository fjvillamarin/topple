class CounterView(BaseView):
    def __init__(self, count):
        super().__init__()
        self.count = count

    def _render(self) -> Element:
        return el("div", "")

