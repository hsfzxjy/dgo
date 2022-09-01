part of dgo;

const _maxDartCallbackId = 1 << 32;

var _dartCallbackId = 0;
final _dartCallbackMap = <int, Function>{};

int _dartCallbackPend(Function fn) {
  _dartCallbackId++;
  if (_dartCallbackId >= _maxDartCallbackId) _dartCallbackId = 0;
  final nextId = _dartCallbackId;
  if (_dartCallbackMap.containsKey(nextId)) {
    throw 'dgo:dart too many dart callbacks pending';
  }
  _dartCallbackMap[nextId] = fn;
  return nextId;
}

void _dartCallbackHandle(List objs) {
  if (objs.isEmpty) throw 'dgo:dart empty argument array';

  final int dcb = objs[0];
  final cf = CallbackFlag._(dcb);
  final id = dcb & (_maxDartCallbackId - 1);

  final Function? fn;
  if (cf.hasPop) {
    fn = _dartCallbackMap.remove(id);
  } else {
    fn = _dartCallbackMap[id];
  }

  if (fn == null) {
    throw 'dgo:dart dart callback not exist, id=$id';
  }

  final args = [];

  if (cf.hasWithCode) args.add(cf);

  if (cf.hasFast) {
    if (objs.length > 1) {
      throw 'dgo:dart dart callback with fastKind != none should have no arguments';
    }
    switch (cf.fastKind) {
      case FastKind.nil:
        args.add(null);
        break;
      case FastKind.yes:
        args.add(true);
        break;
      case FastKind.no:
        args.add(false);
        break;
      default:
    }
  } else {
    args.addAll(objs.sublist(1));
  }

  if (cf.hasPackArray) {
    fn(args);
  } else {
    Function.apply(fn, args);
  }
}
