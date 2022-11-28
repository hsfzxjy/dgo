part of 'generator.dart';

class FileSet {
  final _map = <DartFileUri, File>{};

  File operator [](DartFileUri fileUri) {
    return _map.putIfAbsent(fileUri, () => File(fileUri));
  }

  Future<void> save() async {
    for (final f in _map.values) {
      await f.save();
    }
  }
}
