#! /usr/bin/python

## TODO: rewrite in Go

import sys
import os
import yaml
import json
import argparse

PROJ_PATH = os.path.abspath(__file__)

def msg(message):
    print(message)

## shell


def activate_virtualenv(config, ctx_value, env, args):
    ve = os.path.abspath(ctx_value)
    env['PATH'] = '{}:{}'.format(os.path.join(ve, 'bin'), env['PATH'])
    if 'PYTHONHOME' in env:
        del env['PYTHONHOME']
    env['VIRTUAL_ENV'] = ve
    # TODO: doesn't work
    #env['PS1'] = '({}) {}'.format(os.path.basename(ve), env.get('PS1', '$'))
    return env, args

context_handlers = {
    'activate_virtualenv': activate_virtualenv,
}


def do_shell(config, ctx):
    msg("starting shell")
    env = os.environ.copy()
    args = [env.get('SHELL', '/bin/sh')]

    for c in ctx:
        name, ctx_value = c.items()[0]
        msg("applying {}: {!r}".format(name, ctx_value))
        env, args = context_handlers[name](config, ctx_value, env, args)
    os.execvpe(args[0], args, env)


def _local_reexec(ctx, elt, path):
    # re-run proj within the new environment
    args = [sys.executable, PROJ_PATH]

    # fork off a process to write the config to the process we exec
    r, w = os.pipe()
    if os.fork() == 0:
        os.close(r)
        json.dump(ctx, os.fdopen(w, "w"))
        sys.exit(0)

    os.close(w)
    args.extend(['--cfd', str(r)])
    if path:
        args.append(path)
    os.execv(sys.executable, args)

## context config


context_config_handlers = {
}

def handle_context_config(ctx_cfg):
    if not isinstance(ctx_cfg, dict) or len(ctx_cfg) != 1:
        msg("** invalid context config")

    name, args = ctx_cfg.items()[0]
    fn = context_config_handlers.get(name, lambda n, a: [{n: a}])
    return fn(name, args)

## traversal


def child_cd(config, ctx, elt, path, arg):
    if not isinstance(arg, dict):
        arg = {'dir': arg}
    if 'dir' not in arg:
        msg("** no directory specified for 'cd' in {}".format(config['__file__']))
    os.chdir(arg['dir'])

    _local_reexec(ctx, elt, path)


child_methods = {
    'cd': child_cd,
}

def start_child(config, ctx, elt, path):
    msg("entering " + elt)
    cfg_file = config['__file__']

    if elt not in config['children']:
        msg("** no such child project '{}' in {}".format(elt, cfg_file))
        return

    child = config['children'][elt]
    if len(child) != 1:
        msg("** malformed child project '{}' in {}".format(elt, cfg_file))
    child_method, method_arg = child.items()[0]

    if child_method not in child_methods:
        msg("** no such child method '{}' in {}".format(child_method, cfg_file))

    child_methods[child_method](config, ctx, elt, path, method_arg)

## config management


def get_config(args):
    fn = os.path.expanduser('~/.projrc.yml')
    if os.path.exists(fn):
        rc_config = yaml.load(open(fn))
    else:
        rc_config = {}

    dirname = os.path.basename(os.path.abspath('.'))
    for filename in ['.proj.yml', os.path.join('..', "{}-proj.yml".format(dirname))]:
        if os.path.exists(filename):
            env_config = yaml.load(open(filename))
            break
    else:
        env_config = {}

    rc_config.update({
        '__file__': filename,
        'children': env_config.get('children', {}),
        'context': env_config.get('context', [])
    })

    return rc_config


def load_context(args):
    if args.cfd:
        return json.load(os.fdopen(args.cfd))
    else:
        return []


def main(argv):
    parser = argparse.ArgumentParser(description='Set up shell environments for projects')
    parser.add_argument('--cfd', type=int)
    parser.add_argument('path', nargs='?', default=None)
    # TODO: allow --env-config so callers can set config

    args = parser.parse_args(argv[1:])

    config = get_config(args)

    ctx = load_context(args)
    for ctx_cfg in config['context']:
        ctx.extend(handle_context_config(ctx_cfg))
    
    if not args.path:
        do_shell(config, ctx)
    else:
        if '/' in args.path:
            elt, path = args.path.split('/', 1)
        else:
            elt, path = args.path, None
        start_child(config, ctx, elt, path)


if __name__ == "__main__":
    main(sys.argv)
