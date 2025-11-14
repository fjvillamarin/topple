from topple.psx import BaseView, Element, el, escape, fragment
class Greeting(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        return el("h1", f"Hello, {escape(self.name)}!")

class Profile(BaseView):
    def __init__(self, name: str, title: str):
        super().__init__()
        self.name = name
        self.title = title

    def _render(self) -> Element:
        return el("div", [el("h2", escape(self.name)), el("p", escape(self.title))])

class Dashboard(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        name = "Dashboard"
        title = "Admin Panel"
        return el("div", [el("h3", escape(name)), el("span", escape(title))])

