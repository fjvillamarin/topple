from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
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
        _chain_tmp_1 = filter_adults(self.users)
        adults = _chain_tmp_1
        _root_children_1000.append(el("div", [el("h1", "User Management"), el("p", f"f"Total users:{escape(len(self.users))}, Adults:{escape(len(adults))}""), el("h2", "All Users"), ""]))
        return fragment(_root_children_1000)

