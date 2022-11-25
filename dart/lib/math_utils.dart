part of dgo;

final _canonicalNaN = (ByteData(8)
      ..setUint32(0, 0xFFFFFFFF, Endian.big)
      ..setUint32(4, 0xFFFFFFFF, Endian.big))
    .getFloat64(0, Endian.big);

final _converterBuffer = ByteData(8);

extension _NumBitsExt on num {
  int get bitsAsUint64 {
    num x = this;
    if (x is int) return x;
    return (_converterBuffer..setFloat64(0, x as double, Endian.big))
        .getUint64(0, Endian.big);
  }

  double get bitsAsFloat64 {
    num x = this;
    if (x is double) return x;
    return (_converterBuffer..setUint64(0, x as int, Endian.big))
        .getFloat64(0, Endian.big);
  }
}

extension _CanonicalNaNExt on double {
  double get canonicalized => isNaN ? _canonicalNaN : this;
}

extension _HexExt on int {
  String get hexUint64 => toRadixString(16).padLeft(16, '0').toUpperCase();
}
