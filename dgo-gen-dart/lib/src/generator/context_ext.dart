part of 'generator.dart';

typedef BlockFunction = void Function();

extension GeneratorContextExt on GeneratorContext {
  void str(Object? content) => this..buffer.write(content);
  void sln([Object? content = '']) => this..buffer.writeln(content);
  void if_(bool cond, BlockFunction ifClause, {BlockFunction? else_}) {
    if (cond) {
      ifClause();
    } else if (else_ != null) {
      else_();
    }
  }

  void pipe(void input) {}
  void then(BlockFunction f) => f();

  void for_<T>(Iterable<T> iter, void Function(T) f) => iter.forEach(f);

  void scope(Map<GeneratorSymbol, String> symbolMap, Function fn) {
    final newContext = GeneratorContext(buffer, importer)
      .._symbols.addAll(_symbols)
      .._symbols.addAll(symbolMap)
      .._usedNames.addAll(_usedNames.followedBy(symbolMap.values));
    GeneratorContext._stack.add(newContext);
    try {
      fn();
    } finally {
      GeneratorContext._stack.removeLast();
    }
  }
}
