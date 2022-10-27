part of 'generator.dart';

class FileSet {
  final _map = <DartFileUri, File>{};

  File operator [](DartFileUri fileUri) {
    return _map.putIfAbsent(fileUri, () => File(fileUri));
  }

  void save(String destDir) {
    for (var f in _map.values) {
      f.save(destDir);
    }
  }
}
