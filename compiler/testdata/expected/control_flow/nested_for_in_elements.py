from topple.psx import BaseView, Element, el, escape, fragment
class Table(BaseView):
    def __init__(self, rows: list):
        super().__init__()
        self.rows = rows

    def _render(self) -> Element:
        _root_children_1000 = []
        _table_children_2000 = []
        for row in self.rows:
            _tr_children_3000 = []
            for cell in row:
                _tr_children_3000.append(el("td", escape(cell)))
            _table_children_2000.append(el("tr", _tr_children_3000))
        _root_children_1000.append(el("table", _table_children_2000))
        return fragment(_root_children_1000)

