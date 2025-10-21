from psx.runtime import BaseView, Element, el, escape, fragment
class HelloWorld(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", "Hello World")

