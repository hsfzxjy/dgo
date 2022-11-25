part of dgo;

class DartCallback implements _Serializable {
  final int _id;
  final DgoPort _port;
  const DartCallback._(this._id, this._port);

  @override
  int get _payload => _id;

  @override
  final _kind = _SpecialIntKind.dartCalback;

  void remove() {
    _port._callbacks.remove(_id);
  }

  @visibleForTesting
  int get id => _id;

  @visibleForTesting
  bool get exists => _port._callbacks.containsKey(_id);
}

class DartFutureCallback extends DartCallback {
  const DartFutureCallback._(int id, DgoPort dgoPort) : super._(id, dgoPort);

  @override
  // ignore: overridden_fields
  final _kind = _SpecialIntKind.dartFutureCallback;
}

class DartStreamCallback extends DartCallback {
  const DartStreamCallback._(int id, DgoPort dgoPort) : super._(id, dgoPort);

  @override
  // ignore: overridden_fields
  final _kind = _SpecialIntKind.dartStreamCallback;
}

class _InvokingDartCallback extends _Handlable {
  final _CallbackId _id;
  final CallbackFlag _flag;
  final DgoPort _port;

  _InvokingDartCallback._(int value, this._port)
      : _id = value & _callbackIdMask,
        _flag = CallbackFlag._(value);

@override
  String toString() => '$runtimeType[id=$_id, port=$_port]';

  @override
  void _handleObjects(Iterable objs) {
    if (_flag.hasFallible) {
      try {
        _handleObjectsFallible(objs);
      } catch (e, st) {
        _logger.warning('error caught in $this', e, st);
      }
    } else {
      _handleObjectsFallible(objs);
    }
  }

  void _handleObjectsFallible(Iterable objs) {
    final Function? fn;
    {
      final callbacks = _port._callbacks;
      if (_flag.hasPop) {
        fn = callbacks.remove(_id);
      } else {
        fn = callbacks[_id];
      }
    }

    if (fn == null) {
      throw 'dgo:dart: callback not exist, $this';
    }

    var args = Iterable.empty();

    if (_flag.hasWithContext) {
      args = _OneShot(InvokeContext(_flag, _port));
    }

    if (_flag.hasFast) {
      if (objs.isNotEmpty) {
        throw 'dgo:dart: expect zero argument when called with FAST flag, $this';
      }
      switch (_flag.fastKind) {
        case FastKind.nil:
          args = args.followedBy(const _OneShot(null));
          break;
        case FastKind.yes:
          args = args.followedBy(const _OneShot(true));
          break;
        case FastKind.no:
          args = args.followedBy(const _OneShot(false));
          break;
        default:
      }
    } else {
      args = args.followedBy(objs);
    }

    if (_flag.hasPackArray) {
      fn(args.toList());
    } else {
      Function.apply(fn, args.toList());
    }
  }
}
