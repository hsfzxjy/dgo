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

    file.if_(
      ir is OpStruct,
      () => file
        ..for_(
            (ir as OpStruct).fields.values,
            (field) => file.writeln(
                '''final ${field.term.dartType(file.importer)} ${field.name};'''))
        ..write('const $entryName(')
        ..writeln(
            ir.fields.values.map((field) => 'this.${field.name}').join(','))
        ..writeln(');'),
      () => file
        ..writeln('final ${ir.dartType(file.importer)} \$inner;')
        ..writeln('const $entryName(this.\$inner);'),
    );

    var ctx = file.asGeneratorContext().withSymbols(
        {vArgs: '\$args', vIndex: '\$startIndex', vHolder: '\$instance'});
    file
      ..writeln()
      ..writeln(
          'static DgoTypeLoadResult<$entryName> \$dgoLoad(List<dynamic> ${ctx[vArgs]}, int ${ctx[vIndex]}){')
      ..writeln('${ir.dartType(file.importer)} ${ctx[vHolder]};')
      ..pipe(ir.writeSnippet$dgoLoad(ctx))
      ..if_(
        ir is OpStruct,
        () => file.writeln(
            'return DgoTypeLoadResult(${ctx[vIndex]}, ${ctx[vHolder]});'),
        () => file.writeln(
            'return DgoTypeLoadResult(${ctx[vIndex]}, $entryName(${ctx[vHolder]}));'),
      )
      ..writeln('}');

    for (final methodSpec in entry['Methods'] ?? []) {
      Method.fromMap(ir, methodSpec).writeSnippet(ctx);
    }

    file
      ..writeln()
      ..writeln('int \$dgoStore(List<dynamic> ${ctx[vArgs]}, int ${ctx[vIndex]}) {')
      ..writeln('final ${ctx[vHolder]} = this;')
      ..pipe(ir.writeSnippet$dgoStore(ctx))
      ..writeln('return ${ctx[vIndex]};')
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
