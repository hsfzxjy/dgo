part of 'ir.dart';

final _irBuilders = [
  OpArray,
  OpBasic,
  OpCoerce,
  OpPtrTo,
  OpField,
  OpStruct,
  OpOptional,
  OpSlice,
  OpMap,
  OpPinToken,
].map((type) {
  final mirror = reflectClass(type);
  final className = type.toString();
  return MapEntry(
    className.substring(2),
    (JsonMap m) => mirror.newInstance(Symbol('fromMap'), [m]).reflectee as IR,
  );
}).asMap();

IR _buildIR(JsonMap m) {
  final String opName = m['Op'];
  final builder = _irBuilders[opName];
  if (builder == null) {
    throw 'unknown IR type: $opName';
  }
  return builder(m);
}

IR? _buildIRNull(JsonMap? m) => m == null ? null : _buildIR(m);
