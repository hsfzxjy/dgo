part of dgo;

typedef NativePortId = int;

abstract class DgoPortLike {
  DgoPort get _port;
}

mixin _PortMixin on DgoPortLike {
  DartCallback pend(Function function) =>
      DartCallback._(_port._pendFunction(function), _port);

  DartFutureCallback pendFuture(Completer completer) {
    void callback(InvokeContext context, dynamic value) {
      final flag = context.flag;
      if (flag.isBitSet(_cfFutureResolve)) {
        completer.complete(value);
      } else {
        completer.completeError(value);
      }
    }

    return DartFutureCallback._(_port._pendFunction(callback), _port);
  }

  DartStreamCallback pendStream(StreamSink sink) {
    void callback(InvokeContext context, dynamic value) {
      final flag = context.flag;
      if (flag.fastKind == FastKind.nil) {
        sink.close();
        return;
      }
      if (flag.isBitSet(_cfStreamError)) {
        sink.addError(value);
      } else {
        sink.add(value);
      }
    }

    return DartStreamCallback._(_port._pendFunction(callback), _port);
  }

  FutureOr<void> close() => _port._close();
}

typedef DgoPortKey = NativePortId;

enum DgoPortState {
  ready,
  closing,
  closed,
}

class DgoPort extends DgoPortLike with _PortMixin, _CallbacksMixin {
  @override
  DgoPort get _port => this;

  final binding.LibDgo _lib;
  final ReceivePort _receivePort;
  late final SendPort _sendPort;

  final _initCompleter = Completer<void>();
  final _closeCompleter = Completer<void>();

  var _state = DgoPortState.ready;
  DgoPortState get state => _state;

  DgoPortKey get key => _sendPort.nativePort;
  bool get isDefault => this == dgo._defaultPort;

  static Future<DgoPort> _build(String receivePortName, DynamicLibrary dylib,
      {required bool isDefault}) {
    if (isDefault && dgo._defaultPort != null) {
      throw 'dgo:dart: default port is already set';
    }

    final receivePort = ReceivePort(receivePortName);
    final lib = binding.LibDgo(dylib);
    final port = DgoPort._(receivePort, lib);

    if (isDefault) dgo._defaultPort = port;

    lib.dgo_InitPort(
      NativeApi.initializeApiDLData,
      receivePort.sendPort.nativePort,
      isDefault,
    );

    receivePort.listen(
      port._handleMessage,
      onError: (_) => port._close(),
      onDone: port._close,
    );

    return port._initCompleter.future.then((_) => port);
  }

  DgoPort._(this._receivePort, this._lib);

  FutureOr<void> _close() {
    if (_state != DgoPortState.ready) return null;
    _state = DgoPortState.closing;
    _sendPort.send(null);
    return _closeCompleter.future.timeout(
      Duration(seconds: 1),
      onTimeout: () {
        _logger.warning('dgo:dart: timeout on closing port=$this');
      },
    ).whenComplete(() {
      _state = DgoPortState.closed;
      if (isDefault) dgo._defaultPort = null;
      _receivePort.close();
    });
  }

  void _postInt(int arg) => _sendPort.send(arg);

  void _postList(Iterable args) => _sendPort.send(args.map((arg) {
        if (arg is _Serializable) {
          return arg.serialize(asDouble: true);
        } else if (arg is double) {
          return arg.canonicalized;
        } else {
          return arg;
        }
      }).toList(growable: false));

  @override
  String toString() => 'DgoPort['
      'S=${_sendPort.nativePort.hexUint64}, '
      'R=${_receivePort.sendPort.nativePort.hexUint64}]';

  void _handleMessage(dynamic message) {
    Iterable args;
    num firstArg;

    if (message is SendPort) {
      _sendPort = message;
      _initCompleter.complete();
      return;
    } else if (message == null) {
      _closeCompleter.complete();
      return;
    } else if (message is List) {
      firstArg = message[0];
      args = message.skip(1);
    } else if (message is int) {
      if (message == 0) return;
      firstArg = message;
      args = Iterable.empty();
    } else {
      throw 'dgo:dart: unsupported message=$message';
    }

    final handler = firstArg.parseHandlable(this);
    if (handler == null) {
      throw 'dgo:dart: cannot parse the first argument $firstArg as handlable';
    }
    handler._handleObjects(args.map((arg) {
      if (arg is! double) return arg;
      return arg.parse(this) ?? arg;
    }));
  }
}

mixin _CallbacksMixin {
  final _callbacks = <_CallbackId, Function>{};

  _CallbackId _nextId = 0;
  static const _maxCallbacks = 1 << 32;
  _CallbackId _getNextId() {
    _nextId++;
    if (_nextId >= _maxCallbacks) _nextId = 0;
    return _nextId;
  }

  int _pendFunction(Function function) {
    final nextId = _getNextId();
    if (_callbacks.containsKey(nextId)) {
      throw 'dgo:dart: too many dart callbacks pending';
    }
    _callbacks[nextId] = function;
    return nextId;
  }
}
