part of 'generator.dart';

class Importer {
  final DartFileUri currentFileUri;
  final imports = <ImportPath, ImportAlias>{};
  var _num = 0;
  Importer(this.currentFileUri);

  ImportAlias import(DartFileUri dartFileUri) {
    if (currentFileUri == dartFileUri) return '';
    return imports.putIfAbsent(dartFileUri.relativeTo(currentFileUri.goMod),
        () {
      _num++;
      return '\$i$_num';
    });
  }

  void import3Party(String pkgName, {String alias = ''}) {
    imports[pkgName] = alias;
  }

  String qualifyUri(EntryUri entry) {
    final alias = import(entry.goMod.dartModFile);
    final name = entry.name;
    if (alias.isEmpty) {
      return name;
    } else {
      return '$alias.$name';
    }
  }

  void writeTo(StringSink sink) {
    for (final imp in imports.entries) {
      sink.write('''import '${imp.key}' ''');
      if (imp.value.isNotEmpty) {
        sink.write('as ${imp.value}');
      }
      sink.writeln(';');
    }
  }
}
