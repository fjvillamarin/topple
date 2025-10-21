from topple.psx import BaseView, Element, el, escape, fragment
class SimpleView(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", "Simple view content")

