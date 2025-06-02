# Parser Implementation Bugs

This document tracks confirmed parser implementation bugs that need to be fixed in the parser code.

## Operator Precedence Issues

### 1. Bitwise OR vs Comparison Precedence
**File**: `comparisons_test.go:121`  
**Test**: `TestBitwiseOr/mixed_with_comparison`  
**Input**: `x | y == 0`  
**Expected**: `x | (y == 0)` - `==` should have higher precedence  
**Actual**: `(x | y) == 0` - `|` is being given higher precedence  
**Python Behavior**: `==` has higher precedence than `|`  
**Impact**: High - incorrect operator precedence affects expression evaluation  

**Root Cause Analysis**: The operator precedence hierarchy in the parser is implemented correctly in principle, but the test results show that bitwise OR (`|`) is being parsed with higher precedence than equality (`==`). This indicates an issue in the precedence chain where `bitwiseOr()` is being called before `comparison()`, when it should be the reverse.

**Parser Location**: `compiler/parser/bitwise.go` and `compiler/parser/comparisons.go`
**Fix Required**: Reorder the expression precedence chain to ensure comparison operators are parsed before bitwise operators.  

## Error Validation Issues

### 2. Multiple Starred Expressions in For Loop Targets
**File**: `for_test.go:333`  
**Test**: `TestStarTargets/multiple_stars_in_same_level`  
**Input**: `for *a, *b in items:`  
**Expected**: Should raise an error - multiple starred expressions at same level not allowed  
**Actual**: Parser accepts this syntax without error  
**Python Behavior**: `SyntaxError: multiple starred expressions in assignment`  
**Impact**: Medium - allows invalid Python syntax to pass  

**Root Cause Analysis**: The issue is in `compiler/parser/targets.go` in the `starTargets()` function (lines 486-491). The parser correctly parses individual starred expressions but doesn't validate that only one starred expression is allowed per assignment level. Additionally, the test shows that method calls in targets are not being parsed correctly (getting 2 targets instead of expected 3).

**Parser Location**: `compiler/parser/targets.go:486-491` and `compiler/parser/for.go`
**Fix Required**: Add validation in the target parsing logic to track and reject multiple starred expressions at the same nesting level.  

### 3. Incomplete "not in" Expression
**File**: `comparisons_test.go:284`  
**Test**: `TestComparisonErrors/incomplete_not_in`  
**Input**: `x not`  
**Expected**: Should raise an error - incomplete "not in" expression  
**Actual**: Parser accepts this as valid unary NOT expression  
**Python Behavior**: `SyntaxError: invalid syntax`  
**Impact**: Low - edge case error handling  

**Root Cause Analysis**: The parser is correctly parsing `"x"` as a complete expression and ignoring the trailing `"not"` token. The issue is in `compiler/parser/comparisons.go` in the `inversion()` method (lines 97-111). The parser should detect when `not` appears at the end of an expression without a following operand.

**Parser Location**: `compiler/parser/comparisons.go:97-111`
**Fix Required**: Add validation to ensure `not` is always followed by a valid expression, or improve the lexer to detect incomplete `not in` constructs as syntax errors.  

## Priority Assessment

**High Priority**:
- Operator precedence bug (#1) - affects core expression evaluation

**Medium Priority**:  
- Multiple starred expressions validation (#2) - syntax correctness

**Low Priority**:
- Incomplete "not in" validation (#3) - edge case error handling

## F-String Feature Limitations

### 4. Equals Operator in F-String Expressions  
**File**: `fstring_test.go:386`  
**Test**: `TestFStringDebugExpressions/equals_in_expression`  
**Input**: `f"{x == y}"`  
**Expected**: Should parse equality comparison inside f-string  
**Actual**: Parser error: `at '=': expected '}'`  
**Impact**: Low - advanced f-string feature limitation  
**Note**: Parser may not fully support complex expressions with `==` inside f-strings

## Function Parameter Parsing Issues

### 5. Incorrect Parameter Counting in Complex Functions
**File**: `function_test.go:467`  
**Test**: `TestFunctionParameterPatterns/complete_parameter_pattern`  
**Input**: `def func(a, b=2, *args, c, d=4, **kwargs): pass`  
**Expected**: 1 regular parameter, 3 parameters with defaults  
**Actual**: 2 regular parameters, 2 parameters with defaults  
**Impact**: Medium - incorrect parameter categorization  

### 6. Function Name Validation Issues
**File**: `function_test.go:30`  
**Test**: Multiple test cases in `TestFunctionEdgeCases`  
**Expected**: Function name should be extracted correctly  
**Actual**: Function name extraction returning empty string  
**Impact**: Medium - affects function definition parsing  

## Expression Grouping Issues

### 7. List Comprehension Type Detection
**File**: `group_test.go:80`  
**Test**: `TestGroupedExpressions/grouped_list_comprehension`  
**Input**: `(x for x in items)`  
**Expected**: Should detect as list comprehension type  
**Actual**: Incorrectly categorized as regular list  
**Impact**: Low - type detection accuracy  

**Root Cause Analysis**: The parenthesized expression `(x for x in items)` should be parsed as a generator expression, not a list comprehension. The issue is in the AST node type classification where generator expressions in parentheses are being incorrectly identified.

**Parser Location**: `compiler/parser/group.go` and comprehension parsing logic
**Fix Required**: Improve type detection to distinguish between list comprehensions `[x for x in items]` and generator expressions `(x for x in items)`.  

### 8. Starred Expression in Groups
**File**: `group_test.go:594`  
**Test**: `TestGroupExpressionEdgeCases/grouped_starred_expression`  
**Input**: `(*args)`  
**Expected**: Should parse starred expression in parentheses  
**Actual**: Parser error: "starred expression cannot appear in a group"  
**Python Behavior**: `(*args,)` is valid tuple syntax  
**Impact**: Medium - valid Python syntax rejected  

**Root Cause Analysis**: The parser correctly rejects `(*args)` as invalid syntax. However, the test expectation may be wrong - Python actually allows `(*args,)` (tuple with trailing comma) but not `(*args)` (grouped starred expression). The issue is in `compiler/parser/group.go` line 36 where the parser calls `namedExpression()` which eventually rejects standalone starred expressions.

**Parser Location**: `compiler/parser/group.go:36`
**Fix Required**: This may be a test expectation issue rather than a parser bug. Verify if `(*args)` should actually be valid Python syntax or if the test should be `(*args,)`.  

## Import Statement Issues

### 9. Relative Import Dot Counting
**File**: `imports_test.go:81`  
**Test**: Multiple relative import tests  
**Input**: `from ..module import name`  
**Expected**: Should count 2 dots for parent directory  
**Actual**: Only counting 1 dot  
**Impact**: High - affects relative import resolution  

**Root Cause Analysis**: The dot counting logic in `compiler/parser/imports.go` lines 61-69 incorrectly handles consecutive dots. The parser is counting `".."` as 1 dot instead of 2. The issue stems from confusion between `Dot` tokens (`.`) and `Ellipsis` tokens (`...`) in the lexer.

**Parser Location**: `compiler/parser/imports.go:61-69`
**Fix Required**: Fix the lexer to emit separate `Dot` tokens for `".."` and update the parser's dot counting logic to properly handle consecutive `Dot` tokens.  

### 10. Dotted Import Path Parsing
**File**: `imports_test.go:632`  
**Test**: `TestImportEdgeCases/very_long_dotted_import`  
**Input**: Long dotted module paths  
**Expected**: Should parse complex dotted import paths  
**Actual**: Parser error on complex dotted paths  
**Impact**: Medium - limits valid Python import patterns  

**Root Cause Analysis**: The parser fails on complex dotted import paths with error "expected identifier after '.'" The issue is in the dotted name parsing logic where the parser doesn't properly handle very long chains of dotted identifiers.

**Parser Location**: `compiler/parser/imports.go` dotted name parsing
**Fix Required**: Improve the dotted name parsing logic to handle arbitrarily long chains of module.submodule.component paths.  

### 11. Import Alias Validation
**File**: `imports_test.go:622`  
**Test**: `TestImportEdgeCases/invalid_alias_keyword`  
**Expected**: Should reject invalid alias syntax  
**Actual**: Parser accepts invalid alias syntax  
**Impact**: Low - validation completeness  

**Root Cause Analysis**: The parser is accepting invalid alias syntax when it should reject it. The issue is in the import alias parsing logic which doesn't properly validate alias names.

**Parser Location**: `compiler/parser/imports.go` alias parsing
**Fix Required**: Add validation to ensure import aliases follow Python identifier rules and reject invalid alias constructs.  

## Control Flow Issues

### 12. Generator Expressions in If Conditions
**File**: `if_test.go:607`  
**Test**: `TestIfEdgeCases/if_with_comprehension_condition`  
**Input**: `if any(x > 0 for x in values):`  
**Expected**: Should parse generator expression in if condition  
**Actual**: Parser error: "expected ')' after arguments"  
**Impact**: Medium - valid Python syntax not supported  

**Root Cause Analysis**: The parser fails on `"if any(x > 0 for x in values):"` because the comprehension syntax inside function calls isn't properly supported. The issue is in `compiler/parser/if.go` line 17 where `ifStatement()` calls `namedExpression()`, but the expression parser doesn't handle generator expressions in function call contexts.

**Parser Location**: `compiler/parser/if.go:17` and `compiler/parser/arguments.go`
**Fix Required**: Update the `arguments()` parser to properly handle generator expressions without parentheses inside function calls.  

## Updated Priority Assessment

**High Priority**:
- Operator precedence bug (#1) - affects core expression evaluation
- Relative import dot counting (#9) - affects import resolution

**Medium Priority**:  
- Multiple starred expressions validation (#2) - syntax correctness
- Function parameter counting (#5, #6) - function definition parsing
- Starred expressions in groups (#8) - valid syntax rejection
- Dotted import parsing (#10) - import statement completeness
- Generator expressions in conditions (#12) - control flow parsing

**Low Priority**:
- Incomplete "not in" validation (#3) - edge case error handling
- F-string equals operator (#4) - advanced f-string features
- Expression type detection (#7) - categorization accuracy
- Import alias validation (#11) - validation completeness

## Summary of Core Issues

After comprehensive analysis of the failing tests, the parser bugs can be categorized into several systemic issues:

### 1. **Expression vs Statement Context Confusion**
Many failures stem from the parser not properly distinguishing between expression and statement contexts. This affects target parsing, generator expressions in function calls, and method call recognition in assignment contexts.

### 2. **Incomplete Error Handling and Validation**
Several parsers succeed when they should fail on incomplete or invalid syntax:
- Incomplete "not" expressions pass validation
- Multiple starred expressions in assignments aren't caught
- Try statement clause ordering isn't validated
- Import alias validation is missing

### 3. **Backtracking and State Management Issues**
Target parsing has fundamental problems with proper backtracking when alternative parses fail, particularly affecting complex chained expressions like `obj.method()[0].attr`.

### 4. **Token Sequence Handling Problems**
The parser struggles with complex token sequences:
- Multiple dots in relative imports (`..`) are counted as single dots
- Complex dotted import paths fail to parse
- HTML content parsing has token recognition issues

### 5. **Missing Core Features**
- Async view statements are completely unsupported
- Basic E2E compilation pipeline is broken
- Generator expressions in function calls don't parse correctly

### 6. **Lexer-Parser Coordination Issues**
Some bugs stem from misalignment between lexer token generation and parser expectations, particularly around relative imports and HTML content parsing.

## Recommended Fixes (Prioritized)

### **Immediate (Critical)**
1. **Fix E2E compilation pipeline** - investigate and resolve the basic view compilation failures
2. **Add async view support** - extend view statement parsing to handle `async view` syntax

### **High Priority**
3. **Fix operator precedence table** in parser to ensure `==` has higher precedence than `|`
4. **Rewrite target parsing logic** to properly handle complex chained expressions and method calls
5. **Fix relative import dot counting** by correcting lexer token generation for consecutive dots
6. **Fix basic subscript target parsing** to handle simple `arr[0]` patterns

### **Medium Priority**  
7. **Add validation** in for-loop target parsing to reject multiple starred expressions at same level
8. **Enhance dotted import path parsing** for complex module hierarchies
9. **Fix generator expression parsing** in function call contexts
10. **Add try statement clause ordering validation**
11. **Fix star target sequence parsing** to correctly count method calls and attributes

### **Low Priority**
12. **Improve error detection** for incomplete "not in" expressions by checking context
13. **Enhance f-string expression parsing** to support equality operators inside replacement fields
14. **Improve error messages** for import-related failures to be more context-specific
15. **Fix AST node type determination** for complex chained expressions

## Target Parsing Issues

### 13. Complex Target Expression Type Resolution
**File**: `targets_test.go:178`  
**Test**: `TestTargetParsing/mixed_access_pattern`  
**Input**: `obj.method()[0].attr`  
**Expected**: Should result in attribute access AST node  
**Actual**: Results in subscript access AST node  
**Python Behavior**: Final `.attr` should be the outermost operation  
**Impact**: Medium - affects complex target expression parsing accuracy  

**Root Cause Analysis**: The parser is incorrectly identifying the AST node type for complex chained expressions. The final operation `.attr` should make this an attribute access, but it's being classified as a subscript access. This indicates an issue in the AST node type determination logic.

**Parser Location**: `compiler/parser/targets.go` type resolution logic
**Fix Required**: Fix the AST node type determination to correctly identify the outermost operation in chained expressions.  

### 14. Subscript Target Parsing in Various Contexts
**File**: `targets_test.go:240`  
**Test**: `TestTargetParsing/subscript_target`  
**Input**: `arr[0]`  
**Expected**: Should parse as subscript target  
**Actual**: Parser error: "expected '.' or '[' after primary expression"  
**Impact**: High - basic subscript targets not parsing correctly  

**Root Cause Analysis**: Basic subscript syntax `arr[0]` is failing to parse with error "expected '.' or '[' after primary expression". This is a fundamental issue in the target parsing logic where the parser isn't recognizing subscript expressions as valid targets.

**Parser Location**: `compiler/parser/targets.go` subscript parsing
**Fix Required**: Fix the basic subscript target parsing to properly handle `identifier[index]` patterns in assignment contexts.  

### 15. Complex Nested Target Expression Parsing
**Files**: Multiple target test failures  
**Tests**: Various `TestTargetEdgeCases` and `TestTargetComplexNesting`  
**Inputs**: Complex chained expressions like `matrix[row][col:end][index]`  
**Expected**: Should parse complex nested access patterns  
**Actual**: Parser errors on nested subscripts and complex expressions  
**Impact**: High - limits support for complex Python target expressions  

**Root Cause Analysis**: Multiple issues in `compiler/parser/targets.go` affecting complex target parsing:
- Line 188-191: `singleTarget()` method call detection logic is flawed
- Line 505-511: `targetWithStarAtom()` doesn't properly backtrack when parsing fails  
- Method calls like `instance.method()` are incorrectly parsed in target contexts
- Deeply nested subscript chains like `matrix[row][col][index]` fail parsing

**Parser Location**: `compiler/parser/targets.go:188-191, 505-511`
**Fix Required**: Comprehensive rewrite of target parsing logic to properly handle complex chained expressions, method calls, and nested subscripts.  

## Try Statement Validation Issues

### 16. Try Statement Clause Ordering Validation
**File**: `try_test.go:679`  
**Test**: `TestTryCombinations/finally_before_except`  
**Input**: `try: pass finally: pass except: pass`  
**Expected**: Should reject invalid clause ordering  
**Actual**: Parser accepts invalid try statement structure  
**Python Behavior**: `SyntaxError: 'finally' cannot appear before 'except'`  
**Impact**: Medium - allows invalid Python syntax to pass  

**Root Cause Analysis**: The parser in `compiler/parser/try.go` lines 66-74 accepts `finally` before `except` clauses, which violates Python syntax rules. The parser should enforce that `except` clauses come before `finally`, but the current logic allows any order.

**Parser Location**: `compiler/parser/try.go:66-74`
**Fix Required**: Add validation to ensure proper clause ordering: `try` → [`except`]* → [`else`]? → [`finally`]?.  

### 17. Exception Type Validation
**File**: `try_test.go:955`  
**Test**: `TestTryEdgeCases/except_with_invalid_expression`  
**Input**: `except 123:`  
**Expected**: Should reject invalid exception type  
**Actual**: Parser accepts numeric literals as exception types  
**Python Behavior**: `TypeError: catching classes that do not inherit from BaseException`  
**Impact**: Low - runtime vs compile-time validation issue  

**Root Cause Analysis**: The parser accepts `except 123:` without validation. While Python allows any expression in except clauses (the error is raised at runtime), the test expects compile-time validation. This is more of a design decision than a bug.

**Parser Location**: `compiler/parser/try.go` exception type parsing
**Fix Required**: Decide whether to add compile-time validation for exception types or adjust test expectations to match Python's runtime validation behavior.

## Additional Critical Bugs Found

### 18. Async View Statement Support
**File**: `view_test.go:56`  
**Test**: `TestAsyncViewStatements/async_view`  
**Input**: `async view AsyncComponent():`  
**Expected**: Should parse async view statements  
**Actual**: Parser error: `at 'async': unexpected token`  
**Impact**: High - async views are not supported  

**Root Cause Analysis**: The parser in `compiler/parser/view.go` line 10 doesn't handle `async` keywords. The `viewStatement()` function expects `view` as the first token but doesn't recognize `async view` constructions.

**Parser Location**: `compiler/parser/view.go:10`  
**Fix Required**: Add support for parsing `async view` statements by checking for the `async` keyword before `view` and setting appropriate flags in the AST.

### 19. Import Error Message Specificity
**File**: `imports_test.go:626`  
**Test**: `TestImportEdgeCases/invalid_relative_import_syntax`  
**Input**: Various malformed relative imports  
**Expected**: Error messages should mention "relative" imports  
**Actual**: Generic "expected identifier" errors  
**Impact**: Low - error message quality affects developer experience  

**Root Cause Analysis**: The parser provides generic error messages like "at '..': expected identifier" instead of specific relative import error messages. This makes debugging import issues more difficult.

**Parser Location**: `compiler/parser/imports.go` error handling
**Fix Required**: Improve error messages to provide context-specific information for relative import syntax errors.

### 20. Star Target Sequence Parsing  
**File**: `targets_test.go:510`  
**Test**: `TestStarTargetSequences/attribute_and_method_targets`  
**Input**: Target expressions with attributes and method calls  
**Expected**: Should parse 3 targets  
**Actual**: Parser only finds 2 targets  
**Impact**: Medium - affects parsing of complex assignment patterns  

**Root Cause Analysis**: The parser is not correctly counting target expressions that include method calls and attribute access. This suggests the target parsing logic is not properly handling method call expressions in assignment contexts.

**Parser Location**: `compiler/parser/targets.go` sequence parsing  
**Fix Required**: Fix target sequence parsing to properly identify and count method calls and attribute access as distinct targets.

### 21. E2E Compilation Failures
**File**: `e2e_test.go:245`  
**Test**: `TestE2EBasicFeatures/Simple_hello_world_view_with_basic_HTML`  
**Input**: Basic view file with HTML  
**Expected**: Should compile successfully  
**Actual**: Parser error: `at '/': unexpected token`  
**Impact**: Critical - basic functionality broken  

**Root Cause Analysis**: The E2E test shows that basic view compilation is failing with unexpected token errors. This suggests fundamental issues with the view parsing or HTML element parsing logic that prevent even simple views from compiling.

**Parser Location**: View parsing and HTML element parsing  
**Fix Required**: Investigate and fix the core view compilation pipeline to handle basic HTML content within views.  

## Import Statement Edge Cases

### 18. Import Error Message Consistency
**File**: `imports_test.go:626`  
**Test**: `TestImportEdgeCases/invalid_relative_import_syntax`  
**Input**: Various malformed relative imports  
**Expected**: Consistent error messaging for relative import issues  
**Actual**: Generic "expected identifier" instead of specific relative import errors  
**Impact**: Low - error message quality  

## Expression Grouping Edge Cases

### 19. Group Expression Error Handling
**File**: `group_test.go:584`  
**Test**: `TestGroupExpressionEdgeCases/extra_closing_parenthesis`  
**Input**: Expressions with mismatched parentheses  
**Expected**: Should detect and report parentheses mismatch errors  
**Actual**: Parser doesn't consistently catch parentheses errors  
**Impact**: Low - edge case error detection  

## Final Priority Assessment

**Critical Priority**:
- E2E compilation failures (#21) - basic functionality completely broken
- Async view statement support (#18) - core feature missing

**High Priority**:
- Operator precedence bug (#1) - affects core expression evaluation
- Relative import dot counting (#9) - affects import resolution  
- Subscript target parsing (#14) - basic functionality broken
- Complex target expression parsing (#15) - limits language support

**Medium Priority**:  
- Multiple starred expressions validation (#2) - syntax correctness
- Function parameter counting (#5, #6) - function definition parsing
- Starred expressions in groups (#8) - valid syntax rejection
- Dotted import parsing (#10) - import statement completeness
- Generator expressions in conditions (#12) - control flow parsing
- Complex target type resolution (#13) - parsing accuracy
- Try statement clause ordering (#16) - syntax validation
- Star target sequence parsing (#20) - assignment pattern parsing

**Low Priority**:
- Incomplete "not in" validation (#3) - edge case error handling
- F-string equals operator (#4) - advanced f-string features
- Expression type detection (#7) - categorization accuracy
- Import alias validation (#11) - validation completeness
- Exception type validation (#17) - runtime vs compile-time issue
- Import error messaging (#19) - error message quality
- Group expression error handling (#19) - edge case detection

## Additional Target Parsing Edge Cases

The following additional target parsing failures have been identified and are related to the broader target parsing issues (#13-15):

### 22. Additional Target Parsing Edge Cases
**File**: `targets_test.go`  
**Tests**: Multiple additional test cases  
**Examples**:
- `TestTargetParsing/chained_subscript_access#01` - nested subscripts parsing
- `TestTargetParsing/subscript_with_star_atom` - starred expressions in subscripts  
- `TestTargetEdgeCases/deeply_nested_subscript_access` - deep subscript chains
- `TestTargetEdgeCases/complex_mixed_access` - complex chained expressions
- `TestTargetEdgeCases/tuple_with_complex_elements` - tuples with complex targets
- `TestTargetComplexNesting/deeply_nested_attribute_chain` - deep attribute chains
- `TestTargetComplexNesting/nested_subscript_and_slice` - mixed subscript/slice notation
- `TestTargetComplexNesting/parenthesized_complex_target` - parenthesized target expressions
- `TestTargetComplexNesting/list_with_mixed_targets` - lists with mixed target types

**Expected**: Should parse various complex target expressions  
**Actual**: Parser errors on nested subscripts, complex expressions, and mixed access patterns  
**Impact**: Medium - expands the scope of target parsing limitations already documented  

**Root Cause Analysis**: These failures are additional manifestations of the core target parsing issues documented in bugs #13-15. They demonstrate that the target parsing logic has fundamental problems with:
- Backtracking when complex parsing fails
- Proper state management for nested expressions  
- Context handling for parenthesized expressions
- Mixed access pattern recognition (attribute + subscript + method calls)

**Parser Location**: `compiler/parser/targets.go` - same areas as bugs #13-15  
**Fix Required**: These will be resolved by the comprehensive target parsing rewrite mentioned for bugs #13-15.

### 23. Try Statement Clause Ordering Extension  
**File**: `try_test.go`  
**Test**: `TestTryCombinations/finally_before_else`  
**Input**: `try: pass finally: pass else: pass`  
**Expected**: Should reject invalid clause ordering (finally before else)  
**Actual**: Parser accepts invalid try statement structure  
**Impact**: Medium - extends bug #16 to additional clause ordering violations  

**Root Cause Analysis**: Related to bug #16 but covers additional invalid clause orderings beyond just finally/except.

**Parser Location**: `compiler/parser/try.go:66-74` - same as bug #16
**Fix Required**: Will be resolved by the same validation fix as bug #16.