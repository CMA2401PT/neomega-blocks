"""Download explicit versions, generate the official report, and create a new lock.

Python 3.11+ standard library only. Requires a Java runtime compatible with the
requested game version. Never selects 'latest' and never overwrites an existing lock.
"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import urllib.request
import zipfile


def fetch(url: str) -> bytes:
    request = urllib.request.Request(url, headers={'User-Agent': 'MinecraftPreview-BlockSources/1'})
    with urllib.request.urlopen(request, timeout=120) as response:
        return response.read()


def download(path: Path, url: str, digest=None, algorithm='sha256') -> bytes:
    if path.exists():
        data = path.read_bytes()
        if digest and hashlib.new(algorithm, data).hexdigest() == digest:
            return data
    data = fetch(url)
    if digest and hashlib.new(algorithm, data).hexdigest() != digest:
        raise ValueError(f'{algorithm} mismatch: {url}')
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(path.name + '.download')
    temporary.write_bytes(data)
    os.replace(temporary, path)
    return data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--java-version', required=True)
    parser.add_argument('--mappings-commit', required=True, help='full 40-character Git commit')
    parser.add_argument('--geyser-release', required=True, help='explicit generator release tag')
    parser.add_argument('--java', default='java', help='compatible java executable')
    parser.add_argument('--sources', type=Path, default=Path.cwd()/'../../sources')
    parser.add_argument('--lock-output', required=True, type=Path)
    args = parser.parse_args()
    if not re.fullmatch(r'[0-9a-f]{40}', args.mappings_commit):
        parser.error('--mappings-commit must be a full lowercase SHA-1')
    for value in [args.java_version, args.geyser_release]:
        if not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9.+_-]*', value):
            parser.error('version/tag contains unsafe path characters')
    if args.lock_output.exists():
        parser.error('lock output already exists; choose a new candidate file')
    sources = args.sources.resolve()
    sources.mkdir(parents=True, exist_ok=True)
    raw_base = f'https://raw.githubusercontent.com/GeyserMC/mappings/{args.mappings_commit}'
    mappings_dir = sources/f'geyser-{args.mappings_commit}'
    readme = download(mappings_dir/'README.md', raw_base+'/README.md')
    versions = re.search(rb'Minecraft: Java Edition ([^ ]+) and Minecraft: Bedrock Edition ([^\s]+)\.', readme)
    if not versions or versions[1].decode() != args.java_version:
        raise ValueError('Geyser README does not target the requested Java version')
    bedrock_version = versions[2].decode()
    mapping_nbt = download(mappings_dir/'blocks.nbt', raw_base+'/blocks.nbt')
    download(mappings_dir/'LICENSE', raw_base+'/LICENSE')

    release_base = f'https://github.com/GeyserMC/mappings-generator/releases/download/{args.geyser_release}'
    release_dir = sources/f'geyser-release-{args.geyser_release}'
    release = json.loads(download(release_dir/'release.json', release_base+'/release.json'))
    full = [asset for asset in release['downloads'].values() if asset['name'].endswith('-full.zip')]
    if len(full) != 1:
        raise ValueError('Geyser release must contain exactly one full archive')
    asset = full[0]
    if not re.fullmatch(r'[A-Za-z0-9._-]+\.zip', asset['name']):
        raise ValueError('invalid release archive name')
    archive_data = download(release_dir/asset['name'], asset['url'], asset['sha256'])
    with zipfile.ZipFile(io.BytesIO(archive_data)) as archive:
        if archive.read('mappings/blocks.nbt') != mapping_nbt:
            raise ValueError('Geyser release NBT differs from the pinned mappings commit; select a matching release')
        for name in ['data/blocks_debug.json', 'README.md', 'LICENSE']:
            path = release_dir/'full'/name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(archive.read(name))
        if versions[0] not in archive.read('README.md'):
            raise ValueError('Geyser release version declaration differs from pinned mappings')

    manifest = json.loads(fetch('https://piston-meta.mojang.com/mc/game/version_manifest_v2.json'))
    match = next((v for v in manifest['versions'] if v['id'] == args.java_version), None)
    if match is None:
        raise ValueError('requested version not found in official Mojang manifest')
    java_dir = sources/f'java-{args.java_version}'
    version = json.loads(download(java_dir/'version.json', match['url'], match['sha1'], 'sha1'))
    server = version['downloads']['server']
    server_data = download(java_dir/'server.jar', server['url'], server['sha1'], 'sha1')
    with zipfile.ZipFile(io.BytesIO(server_data)) as outer:
        inner_paths = [name for name in outer.namelist() if name.startswith('META-INF/versions/') and name.endswith('.jar')]
        if len(inner_paths) != 1:
            raise ValueError('unsupported server bundle layout')
        with zipfile.ZipFile(io.BytesIO(outer.read(inner_paths[0]))) as inner:
            game_data = inner.read('version.json')
    game_version = json.loads(game_data)
    if game_version['id'] != args.java_version:
        raise ValueError('server game version does not match request')
    (java_dir/'game-version.json').write_bytes(game_data)
    command = [args.java, '-DbundlerMainClass=net.minecraft.data.Main', '-jar', 'server.jar', '--reports', '--output', 'generated']
    print(f'Generating Java {args.java_version} report (requires Java {version["javaVersion"]["majorVersion"]})...', flush=True)
    with (java_dir/'datagen.log').open('w', encoding='utf-8') as log:
        subprocess.run(command, cwd=java_dir, stdout=log, stderr=subprocess.STDOUT, check=True)
    report_path = java_dir/'generated/reports/blocks.json'
    if not report_path.is_file():
        raise ValueError('official data generator did not produce reports/blocks.json')

    revision = args.java_version + '; server SHA-1 ' + server['sha1']
    entries = {
        'blocksReport': (report_path, 'Mojang Java block report', server['url'], revision),
        'blocksNBT': (mappings_dir/'blocks.nbt', 'Geyser typed block mappings', raw_base+'/blocks.nbt', args.mappings_commit),
        'blocksDebug': (release_dir/'full/data/blocks_debug.json', 'Geyser named debug mappings', asset['url'], args.geyser_release),
        'gameVersion': (java_dir/'game-version.json', 'Mojang game version from server JAR', server['url'], revision),
        'mappingsReadme': (mappings_dir/'README.md', 'Geyser version declaration', raw_base+'/README.md', args.mappings_commit),
    }
    lock = {'formatVersion': 1, 'javaVersion': args.java_version, 'bedrockVersion': bedrock_version,
            'dataVersion': game_version['world_version'], 'inputs': {}}
    for key, (path, name, url, source_revision) in entries.items():
        lock['inputs'][key] = {'path': path.relative_to(sources).as_posix(), 'source': {
            'name': name, 'url': url, 'revision': source_revision, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest()}}
    args.lock_output.parent.mkdir(parents=True, exist_ok=True)
    # Exclusive creation protects a lock even if another process created it meanwhile.
    with args.lock_output.open('x', encoding='utf-8', newline='\n') as output:
        json.dump(lock, output, ensure_ascii=False, indent=2)
        output.write('\n')
    (java_dir/'generation.json').write_text(json.dumps({'command': command, 'serverSHA1': server['sha1'],
        'versionManifestURL': match['url'], 'requiredJava': version['javaVersion']['majorVersion'],
        'blocksReportSHA256': lock['inputs']['blocksReport']['source']['sha256']}, indent=2)+'\n', encoding='utf-8')
    print('Candidate lock created:', args.lock_output.resolve())
    print('Validate with go run ./third_party/gen_resource/geyser/cmd/generate -lock <candidate> -sources <sources> -output <candidate snapshot> before publishing a preset.')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('Source preparation failed:', error, file=sys.stderr)
        sys.exit(1)
