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
  bool get isGoDynamic => goSize == -1;
  bool get isGoNotDynamic => !isGoDynamic;
  IR(JsonMap m)
      : dartSize = m['DartSize'],
        goSize = m['GoSize'];
  factory IR.fromJSON(Map m) => _buildIR(m as JsonMap);
  String get dartType;
  String get outerDartType => dartType;

  void writeSnippet$dgoLoad();
  void writeSnippet$dgoStore();
  void writeSnippet$dgoGoSize() {
    if (isGoNotDynamic) {
      ctx.sln('$vSize += $goSize;');
      return;
    }
    _writeSnippet$dgoGoSize();
  }

  void _writeSnippet$dgoGoSize();
}

abstract class Namable extends IR {
  final EntryUri? myUri;
  Namable(JsonMap m)
      : myUri = m.myUri,
        super(m);

  @override
  String get outerDartType =>
      isNamed ? ctx.importer.qualifyUri(myUri!) : dartType;
}

extension NamableExt on Namable {
  bool get isNamed => myUri != null;

  String get _snippetQualifier {
    return isNamed ? '.\$inner' : '';
  }

  String get vHolderQ => '$vHolder$_snippetQualifier';
}

@immutable
class OpSlice extends Namable {
  final IR elem;

  OpSlice.fromMap(JsonMap m)
      : elem = m.getIR('Elem'),
        super(m);

  @override
  String get dartType => '\$core.List<${elem.dartType}>';

  @override
  void writeSnippet$dgoLoad() {
    final vSize2 = vSize.dup;
    ctx
      ..sln('final $vSize2 = $vArgs.current; $vArgs.moveNext();')
      ..sln('$vHolder = \$core.List.generate($vSize2, (_) {')
      ..sln('${elem.dartType} $vHolder;')
      ..scope({}, elem.writeSnippet$dgoLoad)
      ..sln('return $vHolder;')
      ..sln('}, growable: false);');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('$vArgs[$vIndex] = $vHolderQ.length;')
      ..sln('$vIndex++;')
      ..sln('for (final $vElement in $vHolderQ){')
      ..alias({vHolder: vElement}, elem.writeSnippet$dgoStore)
      ..sln('}');
  }

  @override
  void _writeSnippet$dgoGoSize() {
    if (elem.isGoNotDynamic) {
      ctx.sln('$vSize += $vHolderQ.length * ${elem.goSize} + 1;');
    } else {
      final vElement = vHolder.dup;
      ctx
        ..sln('for (final $vElement in $vHolderQ) {')
        ..alias({vHolder: vElement}, elem.writeSnippet$dgoGoSize)
        ..sln('}')
        ..sln('$vSize += 1;');
    }
  }
}

@immutable
class OpMap extends Namable {
  final IR key;
  final IR value;

  OpMap.fromMap(JsonMap m)
      : key = m.getIR('Key'),
        value = m.getIR('Value'),
        super(m);

  @override
  String get dartType => '\$core.Map<${key.dartType}, ${value.dartType}>';

  @override
  void writeSnippet$dgoLoad() {
    final vSize2 = vSize.dup;
    ctx
      ..sln('final $vSize2 = $vArgs.current; $vArgs.moveNext();')
      ..sln('$vHolder = \$core.Map.fromEntries(')
      ..sln('\$core.Iterable.generate($vSize2, (_) {')
      ..sln('${key.dartType} key;')
      ..scope({vHolder: 'key'}, key.writeSnippet$dgoLoad)
      ..sln('${value.dartType} value;')
      ..scope({vHolder: 'value'}, value.writeSnippet$dgoLoad)
      ..sln('return \$core.MapEntry(key, value);')
      ..sln('})')
      ..sln(');');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('$vArgs[$vIndex] = $vHolderQ.length;')
      ..sln('$vIndex++;')
      ..sln('for (final entry in $vHolderQ.entries){')
      ..sln('{final $vElement = entry.key;')
      ..alias({vHolder: vElement}, key.writeSnippet$dgoStore)
      ..sln('}{final $vElement = entry.value;')
      ..alias({vHolder: vElement}, value.writeSnippet$dgoStore)
      ..sln('}}');
  }

  @override
  void _writeSnippet$dgoGoSize() {
    if (value.isGoNotDynamic) {
      ctx
        ..sln('$vSize +=')
        ..sln('$vHolderQ.length * (${key.goSize}+${value.goSize}) + 1;');
    } else {
      final vElement = vHolder.dup;
      ctx
        ..sln('for (final $vElement in $vHolder.values) {')
        ..sln('$vSize += ${key.goSize};')
        ..alias({vHolder: vElement}, value.writeSnippet$dgoGoSize)
        ..sln('}')
        ..sln('$vSize += 1;');
    }
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
  String get dartType => '\$core.List<${elem.dartType}>';

  @override
  void writeSnippet$dgoLoad() {
    ctx
      ..sln('$vHolder = \$core.List.generate($len, (index) {')
      ..sln('${elem.dartType} $vHolder;')
      ..scope({}, elem.writeSnippet$dgoLoad)
      ..sln('return $vHolder;')
      ..sln('}, growable: false);');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('for (final $vElement in $vHolderQ){')
      ..alias({vHolder: vElement}, elem.writeSnippet$dgoStore)
      ..sln('}');
  }

  @override
  void _writeSnippet$dgoGoSize() {
    final vElement = vHolder.dup;
    ctx
      ..sln('for (final $vElement in $vHolderQ) {')
      ..alias({vHolder: vElement}, elem.writeSnippet$dgoGoSize)
      ..sln('}');
  }
}

@immutable
class OpBasic extends Namable {
  final String typeName;

  @override
  String get dartType => '\$core.$_dartType';
  String get _dartType {
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
  void writeSnippet$dgoLoad() {
    ctx
      ..sln('$vHolder = $vArgs.current;')
      ..sln('$vArgs.moveNext();');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx
      ..sln('$vArgs[$vIndex] = $vHolder$_snippetQualifier;')
      ..sln('$vIndex++;');
  }

  @override
  void _writeSnippet$dgoGoSize() {}
}

@immutable
class OpCoerce extends IR {
  final EntryUri ident;

  OpCoerce.fromMap(JsonMap m)
      : ident = EntryUri.fromString(m['Target']['Uri']),
        super(m);

  @override
  String get dartType => currentContext.importer.qualifyUri(ident);

  @override
  void writeSnippet$dgoLoad() {
    ctx.sln('$vHolder = $dartType.\$dgoLoad($vPort, $vArgs);');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx.sln('$vIndex = $vHolder.\$dgoStore($vArgs, $vIndex);');
  }

  @override
  void _writeSnippet$dgoGoSize() {
    ctx.sln('$vSize += $vHolder.\$dgoGoSize;');
  }
}

@immutable
class OpPtrTo extends Namable {
  final IR elem;

  OpPtrTo.fromMap(JsonMap m)
      : elem = m.getIR('Elem'),
        super(m);

  @override
  String get dartType => elem.dartType;

  @override
  void writeSnippet$dgoLoad() => elem.writeSnippet$dgoLoad();

  @override
  void writeSnippet$dgoStore() => elem.writeSnippet$dgoStore();

  @override
  void _writeSnippet$dgoGoSize() => elem._writeSnippet$dgoGoSize();
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
  String get dartType => term.dartType;

  @override
  void writeSnippet$dgoLoad() => term.writeSnippet$dgoLoad();

  @override
  void writeSnippet$dgoStore() => term.writeSnippet$dgoStore();

  @override
  void _writeSnippet$dgoGoSize() => term._writeSnippet$dgoGoSize();
}

@immutable
class OpStruct extends Namable {
  final Map<String, OpField> fields;

  OpStruct.fromMap(JsonMap m)
      : fields = (m['Fields'] as List)
            .map((e) => MapEntry(e['Name'] as String, _buildIR(e) as OpField))
            .where((entry) => entry.value.sendToDart)
            .asMap(ordered: true),
        super(m);

  @override
  String get dartType => currentContext.importer.qualifyUri(myUri!);

  Iterable<OpField> get goFields =>
      fields.values.where((field) => field.sendBackToGo);

  @override
  void writeSnippet$dgoLoad() {
    final structName = myUri!.name;
    final vFields = <GeneratorSymbol>[];
    ctx
      ..for_(fields.values, (field) {
        final vField = vHolder.dup;
        vFields.add(vField);
        ctx
          ..sln(' // Loading Field ${field.name}')
          ..sln('${field.dartType} $vField;')
          ..sln('{')
          ..alias({vHolder: vField}, field.writeSnippet$dgoLoad)
          ..sln('}');
      })
      ..sln(' // Constructing instance')
      ..sln('$vHolder = $structName(${vFields.joinComma});');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx.for_(goFields, (field) {
      final vField = vHolder.dup;
      ctx
        ..sln(' // Storing Field ${field.name}')
        ..sln('{')
        ..sln('final $vField = $vHolder.${field.name};')
        ..alias({vHolder: vField}, field.writeSnippet$dgoStore)
        ..sln('}');
    });
  }

  @override
  void _writeSnippet$dgoGoSize() {
    final vField = vHolder.dup;
    ctx.for_(
      goFields,
      (field) => ctx.if_(
        field.isGoNotDynamic,
        () => ctx.sln('$vSize += ${field.goSize};'),
        else_: () => ctx
          ..sln('{ final $vField = $vHolder.${field.name};')
          ..alias({vHolder: vField}, field.writeSnippet$dgoGoSize)
          ..sln('}'),
      ),
    );
  }
}

@immutable
class OpOptional extends IR {
  final IR term;

  OpOptional.fromMap(JsonMap m)
      : term = m.getIR('Term'),
        super(m);

  @override
  String get dartType => '${term.dartType}?';

  @override
  void writeSnippet$dgoLoad() {
    ctx
      ..sln('if ($vArgs.current==null) {')
      ..sln('$vHolder = null;')
      ..sln('$vArgs.moveNext();')
      ..sln('} else {')
      ..then(term.writeSnippet$dgoLoad)
      ..sln('}');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx
      ..sln('if ($vHolder==null) {')
      ..sln('$vArgs[$vIndex] = null;')
      ..sln('$vIndex++;')
      ..sln('} else {')
      ..then(term.writeSnippet$dgoStore)
      ..sln('}');
  }

  @override
  void _writeSnippet$dgoGoSize() {
    ctx
      ..sln('if ($vHolder == null) {')
      ..sln('$vSize += 1;')
      ..sln('} else {')
      ..then(term.writeSnippet$dgoGoSize)
      ..sln('}');
  }
}

@immutable
class OpPinToken extends IR {
  final OpCoerce term;

  OpPinToken.fromMap(JsonMap m)
      : term = m.getIR('Term') as OpCoerce,
        super(m);

  @override
  void _writeSnippet$dgoGoSize() {}

  @override
  String get dartType => '\$dgo.PinToken<${term.dartType}>';

  @override
  void writeSnippet$dgoLoad() {
    ctx.sln('$vHolder = \$dgo.PinToken.\$dgoLoad<${term.dartType}>($vArgs);');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx.sln('$vIndex = $vHolder.\$dgoStore($vArgs, $vIndex);');
  }
}
