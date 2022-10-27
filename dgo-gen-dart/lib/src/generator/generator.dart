// Dart imports:
import 'dart:io' as io;

// Package imports:
import 'package:dart_style/dart_style.dart';
import 'package:meta/meta.dart';
import 'package:path/path.dart' as p;

// Project imports:
import '../ir/ir.dart';
import '../misc.dart';
import '../uri.dart';

part 'importer.dart';
part 'context.dart';
part 'file.dart';
part 'fileset.dart';

class Generator {
  final Map<String, dynamic> ir;
  final String destDir;
  final FileSet fileSet = FileSet();
  Generator(this.ir, this.destDir);

  void _processEntry(Map<String, dynamic> entry) {
    final typeId = entry['TypeId'] as int;
    final ir = IR.fromJSON(entry['Term']);
    assert(ir is Namable && ir.myUri != null);
    final myUri = (ir as Namable).myUri!;
    final file = fileSet[myUri.goMod.dartModFile];
    final entryName = myUri.name;

    file
      ..writeln('@immutable')
      ..writeln('class $entryName {')
      ..writeln('static const typeId = $typeId;');

    if (ir is OpStruct) {
      for (final field in ir.fields.values) {
        file.writeln(
            '''final ${field.term.dartType(file.importer)} ${field.name};''');
      }
      file
        ..write('const $entryName(')
        ..writeln(
            ir.fields.values.map((field) => 'this.${field.name}').join(','))
        ..writeln(');');
    } else {
      file
        ..writeln('final ${ir.dartType(file.importer)} \$inner;')
        ..writeln('const $entryName(this.\$inner);');
    }

    var ctx = file.asGeneratorContext().withSymbols(
        {vArgs: 'args', vIndex: 'startIndex', vHolder: 'instance'});
    file
      ..writeln()
      ..writeln(
          'static DgoTypeLoadResult<$entryName> \$dgoLoad(List<dynamic> args, int startIndex){')
      ..writeln('${ir.dartType(file.importer)} instance;')
      ..pipe(ir.writeSnippet$dgoLoad(ctx))
      ..pipe(() {
        if (ir is OpStruct) {
          file.writeln('return DgoTypeLoadResult(startIndex, instance);');
        } else {
          file.writeln(
              'return DgoTypeLoadResult(startIndex, $entryName(instance));');
        }
      }())
      ..writeln('}');

    file
      ..writeln()
      ..writeln('int \$dgoStore(List<dynamic> args, int startIndex) {')
      ..writeln('final instance = this;')
      ..pipe(ir.writeSnippet$dgoStore(ctx))
      ..writeln('return startIndex;')
      ..writeln('}');

    file.writeln('}');
  }

  void save() {
    for (final entry in ir.values) {
      _processEntry(entry);
    }
    fileSet.save(destDir);
  }
}
