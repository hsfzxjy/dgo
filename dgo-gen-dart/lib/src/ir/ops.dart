part of 'ir.dart';

extension on JsonMap {
  EntryUri? get myUri {
    String? str = this['Ident']?['Uri'];
    if (str != null) {
      return EntryUri.fromString(str);
    } else {
      return null;
    }
  }

  IR getIR(String fieldName) => _buildIR(this[fieldName]);
}

abstract class IR {
  final int dartSize;
  final int goSize;
  IR(JsonMap m)
      : dartSize = m['DartSize'],
        goSize = m['GoSize'];
  factory IR.fromJSON(Map m) => _buildIR(m as JsonMap);
  String dartType(Importer context);
  String outerDartType(Importer context) => dartType(context);

  void writeSnippet$dgoLoad(GeneratorContext ctx);
  void writeSnippet$dgoStore(GeneratorContext ctx);
}

abstract class Namable extends IR {
  final EntryUri? myUri;
  Namable(JsonMap m)
      : myUri = m.myUri,
        super(m);

  @override
  String outerDartType(Importer context) =>
      isNamed ? context.qualifyUri(myUri!) : dartType(context);
}

extension on Namable {
  bool get isNamed => myUri != null;

  String get _snippetQualifier {
    return isNamed ? '.\$inner' : '';
  }
}

@immutable
class OpArray extends Namable {
  final int len;
  final IR elem;

  OpArray.fromMap(JsonMap m)
      : len = m['Len'],
        elem = m.getIR('Elem'),
        super(m);

  @override
  String dartType(Importer context) => 'List<${elem.dartType(context)}>';

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) {
    ctx.buffer
      ..writeln('${ctx[vHolder]} = List.generate($len, (index) {')
      ..writeln('${elem.dartType(ctx.importer)} instance;')
      ..pipe(elem.writeSnippet$dgoLoad(ctx.withSymbol(vHolder, 'instance')))
      ..writeln('return instance;')
      ..writeln('}, growable: false);');
  }

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) {
    final elementName = ctx.pickUnique('\$element');
    ctx.buffer
      ..writeln('for (var i=0;i<$len;i++){')
      ..writeln('final $elementName = ${ctx[vHolder]}$_snippetQualifier[i];')
      ..pipe(elem.writeSnippet$dgoStore(ctx.withSymbol(vHolder, elementName)))
      ..writeln('}');
  }
}

@immutable
class OpBasic extends Namable {
  final String typeName;

  @override
  String dartType(Importer context) {
    switch (typeName) {
      case 'bool':
        return 'bool';
      case 'float32':
      case 'float64':
        return 'double';
      case 'string':
        return 'String';
      default:
        if (typeName.startsWith('int') ||
            typeName.startsWith('uint') ||
            typeName == 'byte') {
          return 'int';
        }
        throw AssertionError('Cannot convert Go type $typeName to Dart type');
    }
  }

  OpBasic.fromMap(JsonMap m)
      : typeName = m['TypeName'],
        super(m);

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) {
    ctx.buffer
      ..writeln('${ctx[vHolder]} = ${ctx[vArgs]}.current;')
      ..writeln('${ctx[vArgs]}.moveNext();');
  }

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) {
    ctx.buffer
      ..writeln(
          '${ctx[vArgs]}[${ctx[vIndex]}] = ${ctx[vHolder]}$_snippetQualifier;')
      ..writeln('${ctx[vIndex]}++;');
  }
}

@immutable
class OpCoerce extends IR {
  final EntryUri ident;

  OpCoerce.fromMap(JsonMap m)
      : ident = EntryUri.fromString(m['Target']['Uri']),
        super(m);

  @override
  String dartType(Importer context) => context.qualifyUri(ident);

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) {
    ctx.buffer.writeln(
        '${ctx[vHolder]} = ${dartType(ctx.importer)}.\$dgoLoad(${ctx[vArgs]});');
  }

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) {
    ctx.buffer.writeln(
        '${ctx[vIndex]} = ${ctx[vHolder]}.\$dgoStore(${ctx[vArgs]}, ${ctx[vIndex]});');
  }
}

@immutable
class OpPtrTo extends Namable {
  final IR elem;

  OpPtrTo.fromMap(JsonMap m)
      : elem = m.getIR('Elem'),
        super(m);

  @override
  String dartType(Importer context) => elem.dartType(context);

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) =>
      elem.writeSnippet$dgoLoad(ctx);

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) =>
      elem.writeSnippet$dgoStore(ctx);
}

@immutable
class OpField extends IR {
  final String _name;
  final IR term;
  final String renameInDart;
  final bool sendToDart;
  final bool sendBackToGo;

  String get name => renameInDart.isNotEmpty ? renameInDart : _name;

  OpField.fromMap(JsonMap m)
      : _name = m['Name'],
        term = m.getIR('Term'),
        renameInDart = m['RenameInDart'],
        sendToDart = m['SendToDart'],
        sendBackToGo = m['SendBackToGo'],
        super(m);

  @override
  String dartType(Importer context) => term.dartType(context);

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) =>
      term.writeSnippet$dgoLoad(ctx);

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) =>
      term.writeSnippet$dgoStore(ctx);
}

@immutable
class OpStruct extends Namable {
  final LinkedHashMap<String, OpField> fields;

  OpStruct.fromMap(JsonMap m)
      : fields = LinkedHashMap.fromEntries((m['Fields'] as List)
            .map((e) => MapEntry(e['Name'] as String, _buildIR(e) as OpField))
            .where((entry) => entry.value.sendToDart)),
        super(m);

  @override
  String dartType(Importer context) {
    return context.qualifyUri(myUri!);
  }

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) {
    final structName = myUri!.name;
    ctx.buffer.writeln('{');
    final vFieldHolders = <String>[];
    for (final field in fields.values) {
      final vFieldHolder = '\$field${field.name}';
      vFieldHolders.add(vFieldHolder);
      ctx.buffer
        ..writeln('${field.dartType(ctx.importer)} $vFieldHolder;')
        ..pipe(
            field.writeSnippet$dgoLoad(ctx.withSymbol(vHolder, vFieldHolder)));
    }
    final constructorArgs = vFieldHolders.join(',');
    ctx.buffer
      ..writeln('${ctx[vHolder]} = $structName($constructorArgs);')
      ..writeln('}');
  }

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) {
    ctx.buffer.writeln('{');
    for (final field in fields.values) {
      if (!field.sendBackToGo) continue;
      ctx.buffer
        ..writeln('{')
        ..writeln('final \$field = ${ctx[vHolder]}.${field.name};')
        ..pipe(field.writeSnippet$dgoStore(ctx.withSymbol(vHolder, '\$field')))
        ..writeln('}');
    }
    ctx.buffer.writeln('}');
  }
}

@immutable
class OpOptional extends IR {
  final IR term;

  OpOptional.fromMap(JsonMap m)
      : term = m.getIR('Term'),
        super(m);

  @override
  String dartType(Importer context) => '${term.dartType(context)}?';

  @override
  void writeSnippet$dgoLoad(GeneratorContext ctx) {
    ctx.buffer
      ..writeln('if (${ctx[vArgs]}.current==null) {')
      ..writeln('${ctx[vHolder]} = null;')
      ..writeln('${ctx[vArgs]}.moveNext();')
      ..writeln('} else {')
      ..pipe(term.writeSnippet$dgoLoad(ctx))
      ..writeln('}');
  }

  @override
  void writeSnippet$dgoStore(GeneratorContext ctx) {
    ctx.buffer
      ..writeln('if (${ctx[vHolder]}==null) {')
      ..writeln('${ctx[vArgs]}[${ctx[vIndex]}] = null;')
      ..writeln('${ctx[vIndex]}++;')
      ..writeln('} else {')
      ..pipe(term.writeSnippet$dgoStore(ctx))
      ..writeln('}');
  }
}
