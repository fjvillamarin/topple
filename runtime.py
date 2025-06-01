# psx_runtime.py

import html
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional, Union

# -----------------------------------------------------------------------------
# 1) Safe escaping function for any value that might be interpolated into HTML
# -----------------------------------------------------------------------------
def escape(raw: Any) -> str:
    """
    Convert raw data into a safely-escaped string for HTML output.
    - If raw is None → return empty string.
    - If raw is a str → escape &, <, >, ", '.
    - Otherwise (int, float, bool, etc.) → convert to str.
    """
    if raw is None:
        return ""
    if isinstance(raw, str):
        return html.escape(raw, quote=True)
    return str(raw)


# -----------------------------------------------------------------------------
# 2) Internal helper to format HTML attributes from a dict
# -----------------------------------------------------------------------------
def _render_attrs(attrs: Dict[str, Any]) -> str:
    """
    Given a dict mapping attribute names to values, produce a string like:
      ' class="btn" disabled id="foo"'
    - Skip any key whose value is False or None.
    - If value is True → render ' key' (boolean attr).
    - Otherwise → render ' key="escaped_value"'.
    """
    pieces: List[str] = []
    for key, val in attrs.items():
        if val is False or val is None:
            continue
        if val is True:
            pieces.append(f" {key}")
        else:
            pieces.append(f' {key}="{escape(val)}"')
    return "".join(pieces)


# -----------------------------------------------------------------------------
# 3) Element class: represents an HTML element (with raw children or nested Elements)
# -----------------------------------------------------------------------------
class Element:
    """
    In-memory representation of an HTML element. When converted to str(),
    it produces a fully-escaped, concatenated HTML string.

    Attributes:
      - tag       : the tag name, e.g. "div", "p", "span"
      - children  : a list of zero or more items, each of which can be:
          • Element (nested HTML),
          • BaseView (another view to render), or
          • str/int/float (text to escape)
      - attrs     : a dict of attribute→value for this tag
      - self_close: if True, renders as a self-closing tag "<tag attrs />"
      - _html_cache: cached HTML string to prevent re-rendering
    """

    def __init__(
        self,
        tag: str,
        children: Union[
            str, "BaseView", "Element", List[Union[str, "BaseView", "Element"]]
        ] = "",
        attrs: Optional[Dict[str, Any]] = None,
        self_close: bool = False,
    ):
        self.tag = tag
        self.self_close = self_close
        self.attrs = attrs or {}
        self._html_cache: Optional[str] = None  # Cache for rendered HTML

        # Normalize children to a list
        if isinstance(children, list):
            self.children = children
        else:
            self.children = [children]

    def __str__(self) -> str:
        """
        Render this Element (and all nested children) as a single HTML string.
        Uses caching to ensure each element is only rendered once.
        """
        # Return cached result if available
        if self._html_cache is not None:
            return self._html_cache

        attrs_str = _render_attrs(self.attrs)

        if self.self_close:
            self._html_cache = f"<{self.tag}{attrs_str} />"
            return self._html_cache

        parts: List[str] = []
        for child in self.children:
            if child is None:
                continue
            if isinstance(child, Element):
                # Removed debug print and use cached rendering
                parts.append(str(child))
            elif isinstance(child, BaseView):
                # child.render() will return str and is now cached in BaseView
                parts.append(child.render())
            else:
                # Literal text or number: escape it
                parts.append(escape(child))
        
        inner_html = "".join(parts)
        self._html_cache = f"<{self.tag}{attrs_str}>{inner_html}</{self.tag}>"
        return self._html_cache


# -----------------------------------------------------------------------------
# 4) FragmentElement class: represents a fragment (no wrapper element)
# -----------------------------------------------------------------------------
class FragmentElement(Element):
    """
    A special Element that renders its children without any wrapper tag.
    This is similar to React fragments - it just concatenates the children
    without adding any containing HTML element.
    """

    def __init__(
        self,
        children: Union[
            str, "BaseView", "Element", List[Union[str, "BaseView", "Element"]]
        ] = "",
    ):
        # Initialize with no tag, no attributes, no self-closing
        super().__init__("", children, {}, False)

    def __str__(self) -> str:
        """
        Render this FragmentElement as just its concatenated children
        without any wrapper tags. Uses caching to prevent re-rendering.
        """
        # Return cached result if available
        if self._html_cache is not None:
            return self._html_cache

        parts: List[str] = []
        for child in self.children:
            if child is None:
                continue
            if isinstance(child, Element):
                parts.append(str(child))
            elif isinstance(child, BaseView):
                # child.render() will return str and is now cached in BaseView
                parts.append(child.render())
            else:
                # Literal text or number: escape it
                parts.append(escape(child))
        
        self._html_cache = "".join(parts)
        return self._html_cache


# -----------------------------------------------------------------------------
# 5) render_child: normalize nested content (BaseView, Element, or literal) to
#    either an Element or a literal string (to be escaped later)
# -----------------------------------------------------------------------------
def render_child(child: Union["BaseView", Element, str, int, float, None]) -> Union[Element, str]:
    """
    If `child` is:
      - None        → return "" (empty string)
      - BaseView    → call child._get_rendered() (which uses caching)
      - Element     → return it unchanged
      - str/int/etc → return raw literal; will be escaped when placed in Element
    """
    if child is None:
        return ""
    if isinstance(child, BaseView):
        # Use cached rendering method to prevent multiple _render() calls
        return child._get_rendered()
    if isinstance(child, Element):
        return child
    # Literal text → return as string; escaping happens when that string is placed inside an Element
    return str(child)


# -----------------------------------------------------------------------------
# 6) el(): a factory function that produces an Element instance
# -----------------------------------------------------------------------------
def el(
    tag: str,
    content: Union[
        str, "BaseView", Element, List[Union[str, "BaseView", Element]]
    ] = "",
    attrs: Optional[Dict[str, Any]] = None,
    self_close: bool = False,
) -> Element:
    """
    Create an Element for the given tag, children, and attributes.

    - tag: the HTML tag name (e.g. "div", "p").
    - content: one of:
        • a string or number (literal text → escaped in Element.__str__),
        • a BaseView instance (nested view),
        • an Element (nested raw HTML),
        • a list mixing any of the above.
    - attrs: optional dict of HTML attribute → value.
    - self_close: if True → render as "<tag attrs />", ignoring children.

    The returned Element, when converted to str(), will produce the final HTML.
    """
    # Normalize content into a list
    if not isinstance(content, list):
        children = [content]
    else:
        children = content

    return Element(tag, children, attrs, self_close)


# -----------------------------------------------------------------------------
# 7) BaseView: minimal abstract base class with caching
# -----------------------------------------------------------------------------
class BaseView(ABC):
    """
    Abstract base class for any PSX view rendered on the server.
    Subclasses must implement _render() → Element or str.

    BaseView.render() wraps _render(), ensuring a final string is returned.
    Uses caching to ensure _render() is only called once per instance.
    """

    @abstractmethod
    def _render(self) -> Union[Element, str]:
        """
        Return either:
          - an Element instance (preferred for nested structure), or
          - a plain string (already-escaped HTML or literal text)

        The generated code should implement _render() rather than render().
        This method will only be called once per instance due to caching.
        """
        ...

    def _get_rendered(self) -> Union[Element, str]:
        """
        Get the rendered result, using cache if available.
        This ensures _render() is only called once per instance.
        """
        # Lazy initialization of cache attribute
        if not hasattr(self, '_render_cache'):
            self._render_cache = None
        
        if self._render_cache is None:
            self._render_cache = self._render()
        return self._render_cache

    def render(self) -> str:
        """
        Calls _render() (with caching), then ensures the result is a string.
        If an Element, convert to str (which escapes and concatenates children appropriately).
        This method uses caching to ensure the final HTML is only generated once.
        """
        # Lazy initialization of cache attribute
        if not hasattr(self, '_html_cache'):
            self._html_cache = None
            
        if self._html_cache is not None:
            return self._html_cache

        result = self._get_rendered()
        if isinstance(result, Element):
            self._html_cache = str(result)
        else:
            self._html_cache = str(result)
        
        return self._html_cache

    def __str__(self) -> str:
        return self.render()


# -----------------------------------------------------------------------------
# 8) fragment(): a function that renders multiple children without a wrapper element
# -----------------------------------------------------------------------------
def fragment(children: List[Union["BaseView", Element, str]]) -> FragmentElement:
    """
    Create a FragmentElement that renders multiple children without a wrapper element.
    This is similar to React fragments - it just concatenates the children
    without adding any containing HTML element.
    
    - children: a list of BaseView instances, Element instances, or strings
    
    Returns a FragmentElement that when rendered produces the concatenated HTML.
    """
    return FragmentElement(children)