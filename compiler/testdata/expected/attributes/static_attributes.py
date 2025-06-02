from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class StaticAttributes(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("img", "", {"src": "/logo.png", "alt": "Company Logo"}), el("input", "", {"type": "text", "placeholder": "Enter your name", "readonly": True}), el("button", "Submit", {"type": "submit", "disabled": True})], {"class": "container", "id": "main-content", "data-testid": "app"})

