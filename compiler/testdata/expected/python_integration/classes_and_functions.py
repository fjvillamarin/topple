from topple.psx import BaseView, Element, el, escape, fragment
from dataclasses import dataclass
from typing import List
@dataclass
class User:
    name: str
    email: str
    age: int
    def is_adult(self) -> bool:
        return self.age >= 18

    @property
    def display_name(self) -> str:
        return f"{self.name} ({self.age})"

def filter_adults(users: List[User]) -> List[User]:
    return [user for user in users if user.is_adult()]

class UserList(BaseView):
    def __init__(self, users: List[User]):
        super().__init__()
        self.users = users

    def _render(self) -> Element:
        _root_children_1000 = []
        adults = filter_adults(self.users)
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "User Management"))
        _div_children_2000.append(el("p", f"Total users: {escape(len(self.users))}, Adults: {escape(len(adults))}"))
        _div_children_2000.append(el("h2", "All Users"))
        for user in self.users:
            _div_children_3000 = []
            _div_children_3000.append(el("h3", escape(user.display_name)))
            _div_children_3000.append(el("p", f"Email: {escape(user.email)}"))
            _div_children_3000.append(el("p", f"Status: {escape("Adult" if user.is_adult() else "Minor")}"))
            _div_children_2000.append(el("div", _div_children_3000, {"class": escape(f"user {"adult" if user.is_adult() else "minor"}")}))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

