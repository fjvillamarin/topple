from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class HelloWorld(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", "Hello World")

