from topple.psx import BaseView, Element, el, escape, fragment
from typing import List, Optional, Dict
class ComplexView(BaseView):
    def __init__(self, title: str, items: List[str]=[], metadata: Optional[Dict[str, str]]=None, *args, **kwargs):
        super().__init__()
        self.title = title
        self.items = items
        self.metadata = metadata
        self.args = args
        self.kwargs = kwargs

    def _render(self) -> Element:
        _root_children_2000 = []
        _div_children_3000 = []
        _div_children_3000.append(el("h1", escape(self.title)))
        _div_children_3000.append(el("p", f"Items count: {escape(len(self.items))}"))
        if self.metadata:
            _div_children_3000.append(el("p", f"Has metadata: {escape(bool(self.metadata))}"))
        _div_children_3000.append(el("p", f"Args: {escape(len(self.args))}"))
        _div_children_3000.append(el("p", f"Kwargs: {escape(len(self.kwargs))}"))
        _root_children_2000.append(el("div", _div_children_3000))
        return fragment(_root_children_2000)

