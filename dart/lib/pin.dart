part of dgo;

abstract class Pinnable extends DgoObject {
  const Pinnable();
}

class PinToken<T extends Pinnable> extends DgoObject {
  final int _version;
  final int _lid;
  final int _key;
  final T $data;

  bool _disposed = false;

  PinToken._(this._version, this._lid, this._key, this.$data);

  T dispose({DgoPort? port}) {
    if (_disposed) {
      throw 'dgo:dart: PinToken.dispose() must be called for only once';
    }
    final lib = (port ?? dgo.defaultPort)._lib;
    if (!_disposed) {
      _disposed = true;
      lib.dgo_DisposeToken(_version, _lid, _key);
    }
    return $data;
  }

  @override
  final $dgoGoSize = 3;

  @override
  int $dgoStore(List<dynamic> args, int index) {
    args[index] = _version;
    args[index + 1] = _lid;
    args[index + 2] = _key;
    return index + 3;
  }

  @override
  String toString() => '$runtimeType($_version, $_lid, $_key)';

  static PinToken<T> $dgoLoad<T extends Pinnable>(Iterator<dynamic> args) {
    final version = args.current;
    args.moveNext();
    final lid = args.current;
    args.moveNext();
    final key = args.current;
    args.moveNext();
    final data = dgo.buildObject<T>(args);
    return PinToken._(version, lid, key, data);
  }
}

U checkTokenValidity<T extends Pinnable, U>(PinToken<T> token, U value) {
  if (token._disposed) {
    throw 'dgo:dart: $token is disposed';
  }
  return value;
}
