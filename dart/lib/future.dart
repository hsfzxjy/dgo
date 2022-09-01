part of dgo;

int _dartCallbackPendCompleter(Completer com) {
  void _callback(CallbackFlag flag, dynamic value) {
    if (flag.isBitSet(_cfFutResolve)) {
      com.complete(value);
    } else {
      com.completeError(value);
    }
  }

  return _dartCallbackPend(_callback);
}
