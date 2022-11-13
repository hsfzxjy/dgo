typedef ImportPath = String;
typedef ImportAlias = String;

extension StringSinkExt on StringSink {
  void pipe(void _) {}
  void then(void Function() f) => f();
  void if_(bool condition, void Function() ifClause,
      [void Function()? elseClause]) {
    if (condition) {
      ifClause();
    } else if (elseClause != null) {
      elseClause();
    }
  }

  void for_<T>(Iterable<T> iter, void Function(T) f) => iter.forEach(f);
}
