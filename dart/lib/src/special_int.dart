part of dgo;

enum _SpecialIntKind {
  dartCalback(0),
  goCallback(1),
  goObject(2),
  goMethod(3),
  dartFutureCallback(4),
  dartStreamCallback(5),
  dartCallbackGroup(6),
  prevseredGoCall(7);

  const _SpecialIntKind(this.value) : assert(value >= 0 && value <= 7);
  factory _SpecialIntKind.fromInt(int value) {
    return _SpecialIntKind.values[value];
  }
  final int value;
}

abstract class _SpecialInt {}

// magic   (13 bits) = 0111_1111_1111_1
// kind    (3 bits)
// payload (48 bits)
abstract class _Serializable extends _SpecialInt {
  int get _payload;
  _SpecialIntKind get _kind;

  static final _payloadMask = 0xFFFFFFFFFFFF; // 48 bits
  static const _magic = 0x0FFF; // 13 bits
}

abstract class _Handlable extends _SpecialInt {
  void _handleObjects(Iterable objs);
}

extension _SerializableExt on _Serializable {
  num serialize({required bool asDouble}) {
    final payload = _payload;
    assert(
      payload & _Serializable._payloadMask == payload,
      'dgo:dart: specialInt overflow (max 48 bits)',
    );

    final value =
        (_Serializable._magic << (64 - 13)) | (_kind.value << 48) | payload;

    return asDouble ? value.bitsAsFloat64 : value;
  }
}

typedef _Builder<R extends _SpecialInt> = R Function(int payload, DgoPort port);

_Builder<_SpecialInt> _kindAsBuilder(_SpecialIntKind kind) {
  switch (kind) {
    case _SpecialIntKind.dartCalback:
      return DartCallback._;
    case _SpecialIntKind.goCallback:
      return GoCallback._;
    default:
      throw 'unreachable';
  }
}

_Builder<_Handlable> _kindAsHandlableBuilder(_SpecialIntKind kind) {
  switch (kind) {
    case _SpecialIntKind.dartCalback:
      return _InvokingDartCallback._;
    case _SpecialIntKind.dartCallbackGroup:
      return _DartCallbackGroup._;
    default:
      throw 'unreachable';
  }
}

extension _NumParseExt on num {
  R? _parse<R extends _SpecialInt>(
    DgoPort port,
    _Builder<R> Function(_SpecialIntKind) getBuilder,
  ) {
    final intValue = bitsAsUint64;

    if ((intValue >>> (64 - 13)) != _Serializable._magic) return null;

    final kind = _SpecialIntKind.fromInt((intValue >>> 48) & 0x7);
    final payload = intValue & ((1 << 48) - 1);

    return getBuilder(kind)(payload, port);
  }

  _SpecialInt? parse(DgoPort port) => _parse(port, _kindAsBuilder);

  _Handlable? parseHandlable(DgoPort port) =>
      _parse(port, _kindAsHandlableBuilder);
}
