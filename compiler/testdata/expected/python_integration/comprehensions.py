from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class Comprehensions(BaseView):
    def __init__(self, numbers: list, items: list):
        super().__init__()
        self.numbers = numbers
        self.items = items

    def _render(self) -> Element:
        _root_children_1000 = []
        _chain_tmp_1 = [x ** 2 for x in numbers if x % 2 == 0]
        squares = _chain_tmp_1
        _chain_tmp_2 = {item["id"]: item["name"] for item in items if "name" in item}
        item_names = _chain_tmp_2
        _chain_tmp_3 = {len(item["name"]) for item in items if "name" in item}
        unique_lengths = _chain_tmp_3
        _root_children_1000.append(el("div", [el("h2", "Comprehensions Demo"), el("h3", "Even squares:"), el("ul", ""), el("h3", "Item mapping:"), el("ul", ""), el("h3", "Unique name lengths:"), el("p", f"f"Lengths:{escape(sorted(unique_lengths))}"")]))
        return fragment(_root_children_1000)

