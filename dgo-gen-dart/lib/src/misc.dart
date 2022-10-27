typedef ImportPath = String;
typedef ImportAlias = String;

extension StringSinkExt on StringSink {
  void pipe(void _) {}
  void then(void Function() f) => f();
}
