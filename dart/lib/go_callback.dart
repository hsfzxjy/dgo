part of dgo;

@immutable
class GoCallback {
  final int _id;
  const GoCallback(this._id);

  CallableGoCallback flag(CallbackFlag cf) =>
      CallableGoCallback(_id | cf._internal, cf);
}

@immutable
class CallableGoCallback {
  final int _id;
  final CallbackFlag _cf;
  const CallableGoCallback(this._id, this._cf);

  void call([List? args]) {
    args ??= [];
    if (_cf.hasFast) {
      if (args.isNotEmpty) {
        throw 'dgo:dart expect no argument when CF_FAST set';
      }
      _sendPort!.send(_id);
    } else {
      _sendPort!.send([_id, ...args]);
    }
  }
}

@immutable
class GoMethod {
  final int funcId;

  const GoMethod(this.funcId);

  void call(List args, int n, int cb) {
    final id = funcId | _cfMethodCall;
    _sendPort!.send([id, ...args, cb]);
  }
}
