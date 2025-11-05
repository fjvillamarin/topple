from topple.psx import BaseView, Element, el, escape, fragment
class MatchView(BaseView):
    def __init__(self, status: str, data: dict):
        super().__init__()
        self.status = status
        self.data = data

    def _render(self) -> Element:
        return el("div", "")

