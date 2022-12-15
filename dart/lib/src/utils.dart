part of dgo;

final _logger = Logger('dgo');

typedef _CallbackId = int;
const _callbackIdMask = (1 << 32) - 1;

class _OneShotIterator<T> extends Iterator<T> {
  @override
  final T current;
  var _visited = false;

  _OneShotIterator(this.current);

  @override
  bool moveNext() {
    if (!_visited) {
      _visited = true;
      return true;
    }
    return false;
  }
}

class _OneShot<T> extends Iterable<T> {
  final T value;
  const _OneShot(this.value);
  @override
  Iterator<T> get iterator => _OneShotIterator(value);
}
