class Card(BaseView):
    def __init__(self, *, children = None):
        super().__init__()
        self.children = children

    def _render(self) -> Element:
        return el("div", render_child(self.children) if self.children is not None else "", {"class": "card"})

