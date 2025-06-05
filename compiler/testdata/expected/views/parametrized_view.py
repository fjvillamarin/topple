from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class Greeting(BaseView):
    def __init__(self, name: str, age: int = 25):
        super().__init__()
        self.name = name
        self.age = age

    def _render(self) -> Element:
        return el("div", [el("h1", f"Hello,{escape(self.name)}!"), el("p", f"You are{escape(self.age)}years old.")])

