part of 'generator.dart';

typedef BlockFunction<R> = R Function();

extension GeneratorContextExt on GeneratorContext {
  void str(Object? content) => this..buffer.write(content);
  void sln([Object? content = '']) => this..buffer.writeln(content);
  void if_(bool cond, Function() ifClause, {Function()? else_}) {
    if (cond) {
      final result = ifClause();
      if (result is String) sln(result);
    } else if (else_ != null) {
      final result = else_();
      if (result is String) sln(result);
    }
  }

  void pipe(void input) {}
  void then(BlockFunction f) => f();

  void for_<T>(Iterable<T> iter, Function(T) f) {
    for (final i in iter) {
      final result = f(i);
      if (result is String) sln(result);
    }
  }

  void scope(Map<GeneratorSymbol, String> symbolMap, Function fn) {
    final newContext = GeneratorContext(buffer, importer)
      .._symbols.addAll(_symbols)
      .._symbols.addAll(symbolMap);
    GeneratorContext._stack.add(newContext);
    try {
      fn();
    } finally {
      GeneratorContext._stack.removeLast();
    }
  }

  void alias(Map<GeneratorSymbol, GeneratorSymbol> symbolMap, Function fn) =>
      scope(symbolMap.mapMap((e) => MapEntry(e.key, this[e.value])), fn);
}
