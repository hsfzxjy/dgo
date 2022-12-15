part of 'ir.dart';

class Param {
  final String name;
  final IR term;

  Param.fromMap(Map m)
      : term = _buildIR(m['Term']),
        name = m['Name'];
}

extension _NullableIRExt on IR? {
  String get dartType => this == null ? 'void' : this!.dartType;
}

extension _IRExt on IR {
  void writeSnippetLoad() {
    if (this is Namable && (this as Namable).isNamed) {
      final myUri = (this as Namable).myUri!;
      ctx
        ..sln('{')
        ..sln('$vHolder = ')
        ..sln(ctx.importer.qualifyUri(myUri))
        ..sln('.\$dgoLoad($vPort, $vArgs);')
        ..sln('}');
    } else {
      writeSnippet$dgoLoad();
    }
  }
}

extension _SumExt<T extends num> on Iterable<T> {
  T sum() => isEmpty ? 0 as T : reduce((a, b) => (a + b) as T);
}

class Method {
  final IR self;
  final String funcName;
  final List<Param> params;
  final int funcId;
  final IR? returnType;
  final bool returnError;

  Method.fromMap(this.self, Map m)
      : funcName = m['Name'],
        funcId = m['FuncId'],
        params =
            (m['Params'] as List).cast<JsonMap>().map(Param.fromMap).toList(),
        returnType = _buildIRNull(m['Return']),
        returnError = m['ReturnError'];

  bool get isParamsGoNotDynamic => params.every((p) => p.term.isGoNotDynamic);

  void _buildSize([int? selfSize]) {
    if (selfSize == -1) throw 'unreachable';
    selfSize ??= self.goSize;
    if (isParamsGoNotDynamic && selfSize != -1) {
      final size = params.map((p) => p.term.goSize).sum() + selfSize;
      ctx.sln('final $vSize = $size;');
      return;
    }

    ctx
      ..sln()
      ..sln('\$core.int $vSize = ${selfSize == -1 ? "\$dgoGoSize" : selfSize};')
      ..for_(
        params,
        (p) => ctx
          ..if_(
            p.term.isGoNotDynamic,
            () => ctx.sln('$vSize += ${p.term.goSize};'),
            else_: () => ctx
              ..sln('{ final $vHolder = ${p.name};')
              ..then(p.term.writeSnippet$dgoGoSize)
              ..sln('}'),
          ),
      )
      ..sln();
  }

  String signature(String funcName) {
    var paramSig =
        params.map((p) => '${p.term.outerDartType} ${p.name}').followedBy([
      '{\$core.Duration? \$timeout, \$dgo.DgoPort? \$port}',
    ]).joinComma;
    return '\$async.Future<${returnType.dartType}> $funcName($paramSig)';
  }

  void writePinTokenSnippet() => _writeSnippet(
      selfSize: 3,
      methodFlag: '\$dgo.GoMethod.pinned',
      storeSelf: () =>
          ctx..sln('$vIndex = _token.\$dgoStore($vArgs, $vIndex);'));

  void writeSnippet() => _writeSnippet(
      storeSelf: () => ctx..sln('$vIndex = \$dgoStore($vArgs, $vIndex);'));

  // TODO: BAD, BAD, refactor this
  void _writeSnippet(
      {int? selfSize,
      String methodFlag = '0',
      required void Function() storeSelf}) {
    ctx
      ..sln('${signature(funcName)} async {')
      ..sln('final \$\$port = $vPort ?? \$dgo.dgo.defaultPort;')
      ..pipe(_buildSize(selfSize))
      ..sln('final $vArgs = \$core.List<\$core.dynamic>'
          '.filled($vSize, null, growable: false);')
      ..sln('var $vIndex = 0;')
      ..then(storeSelf)
      ..for_(
          params,
          (param) => ctx
            ..sln('{')
            ..sln('final $vHolder = ${param.name};')
            ..then(param.term.writeSnippet$dgoStore)
            ..sln('}'))
      ..sln(
        'final \$future = \$dgo.GoMethod($funcId, $methodFlag, \$\$port)'
        '.${returnType == null ? "call" : "callWithResult"}'
        '($vArgs, timeout: \$timeout, hasError: $returnError);',
      )
      ..if_(
        returnType == null,
        () => ctx.sln('return \$future;'),
        else_: () => ctx
          ..sln('{')
          ..sln('final $vArgs = (await \$future).iterator;')
          ..sln('$vArgs.moveNext();')
          ..sln('${returnType!.dartType} $vHolder;')
          ..scope({vPort: '\$\$port'}, returnType!.writeSnippetLoad)
          ..sln('return $vHolder;')
          ..sln('}'),
      )
      ..sln('}');
  }
}
