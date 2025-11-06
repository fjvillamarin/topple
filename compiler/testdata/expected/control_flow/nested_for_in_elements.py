from topple.psx import BaseView, Element, el, escape, fragment
class Table(BaseView):
    def __init__(self, rows: list):
        super().__init__()
        self.rows = rows

    def _render(self) -> Element:
        _root_children_3000 = []
        _table_children_4000 = []
        for row in self.rows:
            _tr_children_5000 = []
            for cell in row:
                _tr_children_5000.append(el("td", escape(cell)))
            _table_children_4000.append(el("tr", _tr_children_5000))
        _root_children_3000.append(el("table", _table_children_4000))
        return fragment(_root_children_3000)

