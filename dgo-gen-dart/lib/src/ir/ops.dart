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
  String get dartType;
  String get outerDartType => dartType;

  void writeSnippet$dgoLoad();
  void writeSnippet$dgoStore();
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

extension on Namable {
  bool get isNamed => myUri != null;

  String get _snippetQualifier {
    return isNamed ? '.\$inner' : '';
  }
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
    ctx
      ..sln('{')
      ..sln('final size = $vArgs.current;$vArgs.moveNext();')
      ..sln('$vHolder = \$core.List.generate(size, (index) {')
      ..sln('${elem.dartType} instance;')
      ..scope({vHolder: 'instance'}, elem.writeSnippet$dgoLoad)
      ..sln('return instance;')
      ..sln('}, growable: false);')
      ..sln('}');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('$vArgs[$vIndex] = $vHolder$_snippetQualifier.length;')
      ..sln('$vIndex++;')
      ..sln('for (var i=0;i<$vHolder$_snippetQualifier.length;i++){')
      ..sln('final $vElement = $vHolder$_snippetQualifier[i];')
      ..alias({vHolder: vElement}, elem.writeSnippet$dgoStore)
      ..sln('}');
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
    ctx
      ..sln('{')
      ..sln('final size = $vArgs.current;$vArgs.moveNext();')
      ..sln(
          '$vHolder = \$core.Map.fromEntries(\$core.Iterable.generate(size, (_) {')
      ..sln('${key.dartType} key;')
      ..scope({vHolder: 'key'}, key.writeSnippet$dgoLoad)
      ..sln('${value.dartType} value;')
      ..scope({vHolder: 'value'}, value.writeSnippet$dgoLoad)
      ..sln('return \$core.MapEntry(key, value);')
      ..sln('}));')
      ..sln('}');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('$vArgs[$vIndex] = $vHolder$_snippetQualifier.length;')
      ..sln('$vIndex++;')
      ..sln('for (final entry in $vHolder$_snippetQualifier.entries){')
      ..sln('{final $vElement = entry.key;')
      ..alias({vHolder: vElement}, key.writeSnippet$dgoStore)
      ..sln('}{final $vElement = entry.value;')
      ..alias({vHolder: vElement}, value.writeSnippet$dgoStore)
      ..sln('}}');
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
      ..sln('${elem.dartType} instance;')
      ..scope({vHolder: 'instance'}, elem.writeSnippet$dgoLoad)
      ..sln('return instance;')
      ..sln('}, growable: false);');
  }

  @override
  void writeSnippet$dgoStore() {
    final vElement = vHolder.dup;
    ctx
      ..sln('for (var i=0;i<$len;i++){')
      ..sln('final $vElement = $vHolder$_snippetQualifier[i];')
      ..alias({vHolder: vElement}, elem.writeSnippet$dgoStore)
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
    ctx.sln('$vHolder = $dartType.\$dgoLoad($vArgs);');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx.sln('$vIndex = $vHolder.\$dgoStore($vArgs, $vIndex);');
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
          ..alias({vHolder: vField}, field.writeSnippet$dgoLoad);
      })
      ..sln('$vHolder = $structName(${vFields.joinComma});');
  }

  @override
  void writeSnippet$dgoStore() {
    ctx.for_(fields.values, (field) {
      if (!field.sendBackToGo) return;
      final vField = vHolder.dup;
      ctx
        ..sln(' // Storing Field ${field.name}')
        ..sln('final $vField = $vHolder.${field.name};')
        ..alias({vHolder: vField}, field.writeSnippet$dgoStore);
    });
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
}
