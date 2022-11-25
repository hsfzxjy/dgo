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
  final int _funcId;
  final DgoPort _port;

  const GoMethod(this._funcId, this._port);

  void call(List args, int n, int cb) {
    _port._postList(<dynamic>[this, cb].followedBy(args));
  }

  @override
  int get _payload => _funcId;

  @override
  final _kind = _SpecialIntKind.goMethod;
}
