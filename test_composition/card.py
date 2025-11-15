from topple.psx import BaseView, Element, el, escape, fragment
class Card(BaseView):
    def __init__(self, title: str):
        super().__init__()
        self.title = title

    def _render(self) -> Element:
        return el("div", el("h2", escape(self.title)), {"class": "card"})

