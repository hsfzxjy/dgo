part of dgo;

@immutable
class GoCallback implements _SpecialInt {
  final _CallbackId _id;
  final DgoPort _port;
  const GoCallback._(this._id, this._port);

  GoCallback.fromRaw(this._id, {DgoPort? dgoPort})
      : _port = dgoPort ?? dgo.defaultPort;

  CallableGoCallback flag(CallbackFlag cf) =>
      CallableGoCallback(_id | cf._internal, _port, cf);
}

@immutable
class CallableGoCallback implements _Serializable {
  final int _id;
  final CallbackFlag _flag;
  final DgoPort _port;
  const CallableGoCallback(this._id, this._port, this._flag);

  @override
  int get _payload => _id | _flag._internal;

  @override
  final _kind = _SpecialIntKind.goCallback;

  void call([List? args]) {
    args ??= [];
    if (_flag.hasFast) {
      if (args.isNotEmpty) {
        throw 'dgo:dart: expect zero argument when called with FAST flag';
      }
      _port._postInt(serialize(asDouble: false) as int);
    } else {
      _port._postList(<dynamic>[this].followedBy(args));
    }
  }
}

@immutable
class GoMethod implements _Serializable {
  static const int pinned = 0x1;
  final int _funcId;
  final int _flag;
  final DgoPort _port;

  const GoMethod(this._funcId, this._flag, this._port);

  @override
  int get _payload => _flag << 32 | _funcId;

  @override
  final _kind = _SpecialIntKind.goMethod;

  Future<T> _call<T>(
      List args, Future<T> future, DartCallback callback, Duration? timeout) {
    _port._postList(<dynamic>[this, callback].followedBy(args));
    if (timeout == null) return future;
    return future.timeout(timeout, onTimeout: () async {
      callback.remove();
      throw 'The Go method invocation fails to respond in $timeout';
    });
  }

  Future<List<dynamic>> callWithResult(List args,
      {required bool hasError, Duration? timeout}) {
    final completer = Completer<List<dynamic>>();
    DartCallback callback;
    if (!hasError) {
      callback = _port.pend(completer.complete);
    } else {
      callback = _port.pend((List<dynamic> args) {
        String? error = args[0];
        if (error == null) {
          completer.complete(args.sublist(1));
        } else {
          completer.completeError(error);
        }
      });
    }
    return _call(args, completer.future, callback, timeout);
  }

  Future<void> call(List args, {required bool hasError, Duration? timeout}) {
    final completer = Completer<void>();
    DartCallback callback;
    if (!hasError) {
      callback = _port.pend(completer.complete);
    } else {
      callback = _port.pend((String? error) {
        if (error == null) {
          completer.complete();
        } else {
          completer.completeError(error);
        }
      });
    }
    return _call(args, completer.future, callback, timeout);
  }
}
