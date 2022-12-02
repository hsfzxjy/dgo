// Dart imports:
import 'dart:collection';

typedef ImportPath = String;
typedef ImportAlias = String;

extension IterableOfMapEntryExt<K, V> on Iterable<MapEntry<K, V>> {
  Map<K, V> asMap({bool ordered = false}) =>
      (ordered ? LinkedHashMap.fromEntries : Map.fromEntries)(this);
}

extension MapMapExt<K, V> on Map<K, V> {
  Map<K2, V2> mapMap<K2, V2>(
    MapEntry<K2, V2> Function(MapEntry<K, V>) f, {
    bool ordered = false,
  }) =>
      entries.map(f).asMap(ordered: ordered);
}
