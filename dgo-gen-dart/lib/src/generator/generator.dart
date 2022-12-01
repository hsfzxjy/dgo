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
part 'config.dart';

class Generator {
  final JsonMap definitions;
  final FileSet fileSet = FileSet();
  Generator(JsonMap payload) : definitions = payload['Definitions'] {
    config = Config(payload['Config'])..validate();
  }

  void _processEntry(JsonMap entry) {
    final typeId = entry['TypeId'] as int;
    final ir = IR.fromJSON(entry['Term']);
    assert(ir is Namable && ir.myUri != null);
    final myUri = (ir as Namable).myUri!;
    final file = fileSet[myUri.goMod.dartModFile];
    final entryName = myUri.name;

    final isEnum = entry['IsEnum'] as bool;
    final enumMembers = entry['EnumMembers'] as List;
    final constructorName = isEnum ? '.of' : '';

    file
      ..if_(isEnum, () {
        file
          ..writeln('enum $entryName {')
          ..writeln(
              enumMembers.map((m) => "${m['Name']}(${m['Value']})").join(', '))
          ..writeln(';')
          ..writeln('factory $entryName.of(\$core.int value) {')
          ..writeln('switch (value) {')
          ..for_(enumMembers, (m) {
            file
              ..writeln("case ${m['Value']}:")
              ..writeln("return ${m['Name']};");
          })
          ..writeln('default:')
          ..writeln("throw 'dgo:dart: cannot convert \$value to $entryName';")
          ..writeln('}')
          ..writeln('}')
          ..writeln();
      }, () {
        file
          ..writeln('@immutable')
          ..writeln('${isEnum ? "enum" : "class"} $entryName {');
      })
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
          'static ${ir.outerDartType(file.importer)} \$dgoLoad(\$core.Iterator<\$core.dynamic> ${ctx[vArgs]}){')
      ..writeln('${ir.dartType(file.importer)} ${ctx[vHolder]};')
      ..pipe(ir.writeSnippet$dgoLoad(ctx))
      ..if_(
        ir is OpStruct,
        () => file.writeln('return ${ctx[vHolder]};'),
        () =>
            file.writeln('return $entryName$constructorName(${ctx[vHolder]});'),
      )
      ..writeln('}');

    for (final methodSpec in entry['Methods'] ?? []) {
      Method.fromMap(ir, methodSpec).writeSnippet(ctx);
    }

    file
      ..writeln()
      ..writeln(
          '\$core.int \$dgoStore(\$core.List<\$core.dynamic> ${ctx[vArgs]}, \$core.int ${ctx[vIndex]}) {')
      ..writeln('final ${ctx[vHolder]} = this;')
      ..pipe(ir.writeSnippet$dgoStore(ctx))
      ..writeln('return ${ctx[vIndex]};')
      ..writeln('}');

    file.writeln('}');

    final rename = entry['Rename'] as String;
    if (rename.isNotEmpty) {
      file
        ..writeln()
        ..writeln('typedef $rename = $entryName;');
    }
  }

  Future<void> _prepareDirectory() async {
    final directory = io.Directory(config.generatedInPath);
    try {
      await directory.delete(recursive: true);
    } on io.FileSystemException catch (e) {
      if (e.osError?.errorCode != 2) {
        rethrow;
      }
    }
    await directory.create(recursive: true);
  }

  Future<void> save() async {
    await _prepareDirectory();
    for (final entry in definitions.values) {
      _processEntry(entry);
    }
    await fileSet.save();
  }
}
