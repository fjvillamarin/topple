from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class Greeting(BaseView):
    def __init__(self, name: str, age: int=25):
        super().__init__()
        self.name = name
        self.age = age

    def _render(self) -> Element:
        return el("div", [el("h1", f"Hello,{escape(self.name)}!"), el("p", f"You are{escape(self.age)}years old.")])

