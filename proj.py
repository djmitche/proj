#! /home/dustin/tmp/ve/bin/python

## TODO: rewrite in Go

import sys
import os
import yaml
import json
import tempfile
import argparse

SUPPORTED_SHELLS = ('bash',)
PROJ_PATH = os.path.abspath(__file__)

def msg(message, fatal=False):
    print(message)
    if fatal:
        sys.exit(1)

## shell


def activate_virtualenv(config, context, ctx_value, env, args):
    ve = os.path.abspath(ctx_value)
    return 'source "{}/bin/activate"'.format(ve)

context_handlers = {
    'activate_virtualenv': activate_virtualenv,
}


def do_shell(config, context):
    msg("starting shell")

    rc_fd, rc_filename = tempfile.mkstemp()

    try:
        args = {
            'bash': ['bash', '--rcfile', rc_filename, '-i'],
        }[context['shell']]

        rc_content = []

        # run the user's shell config first
        rc_content.append({
            'bash': '[ -f ~/.bashrc ] && . ~/.bashrc',
        }[context['shell']])

        env = os.environ.copy()
        for c in context['modifiers']:
            name, ctx_value = c.items()[0]
            msg("applying {}: {!r}".format(name, ctx_value))
            rc = context_handlers[name](config, context, ctx_value, env, args)
            if rc:
                rc_content.append(rc)

        # delete the rc temp file
        rc_content.append({
            'bash': 'rm -f "{}"'.format(rc_filename),
        }[context['shell']])

        # write out the rcfile content and close it
        os.fdopen(rc_fd, "w").write("\n\n".join(rc_content))

        # exec the shell
        os.execvpe(args[0], args, env)
    except Exception:
        os.unlink(rc_filename)
        raise


def _local_reexec(context, elt, path, env_config=None):
    # re-run proj within the new environment
    args = [sys.executable, PROJ_PATH]

    # fork off a process to write the config to the process we exec
    r, w = os.pipe()
    if os.fork() == 0:
        os.close(r)
        json.dump(context, os.fdopen(w, "w"))
        os._exit(0)

    os.close(w)
    args.extend(['--cfd', str(r)])
    if env_config:
        args.extend(['--env-config', env_config])
    args.append(path)
    os.execv(sys.executable, args)

## context config


def ctx_shell(context, name, args):
    shell = args
    if shell not in SUPPORTED_SHELLS:
        msg("** shell {!r} not supported".format(shell), fatal=True)
    context['shell'] = shell

def ctx_default(context, name, args):
    context['modifiers'].append({name: args})

context_config_handlers = {
    'shell': ctx_shell,
}

def update_context(context, ctx_cfg):
    if not isinstance(ctx_cfg, dict) or len(ctx_cfg) != 1:
        msg("** invalid context config", fatal=True)

    name, args = ctx_cfg.items()[0]
    fn = context_config_handlers.get(name, ctx_default)
    return fn(context, name, args)

## traversal


def child_cd(config, context, elt, path, arg):
    if not isinstance(arg, dict):
        arg = {'dir': arg}
    if 'dir' not in arg:
        msg("** no directory specified for 'cd' in {}".format(config['__file__']), fatal=True)
    env_config = None
    if arg.get('config'):
        env_config = os.path.abspath(arg['config'])
    os.chdir(arg['dir'])
    _local_reexec(context, elt, path, env_config=env_config)


child_methods = {
    'cd': child_cd,
}

def start_child(config, context, elt, path):
    msg("entering " + elt)
    cfg_file = config['__file__']

    if elt not in config['children']:
        msg("** no such child project '{}' in {}".format(elt, cfg_file), fatal=True)
        return

    child = config['children'][elt]
    if len(child) != 1:
        msg("** malformed child project '{}' in {}".format(elt, cfg_file), fatal=True)
    child_method, method_arg = child.items()[0]

    if child_method not in child_methods:
        msg("** no such child method '{}' in {}".format(child_method, cfg_file), fatal=True)

    child_methods[child_method](config, context, elt, path, method_arg)

## config management


def get_config(args):
    fn = os.path.expanduser('~/.projrc.yml')
    if os.path.exists(fn):
        rc_config = yaml.load(open(fn))
    else:
        rc_config = {}

    dirname = os.path.basename(os.path.abspath('.'))
    print args.env_config
    if args.env_config:
        if not os.path.exists(args.env_config):
            msg("** explicit config file {!r} not found".format(os.path.abspath(args.env_config)), fatal=True)
        configs = [args.env_config]
    else:
        configs = ['.proj.yml', os.path.join('..', "{}-proj.yml".format(dirname))]
    for filename in configs:
        if os.path.exists(filename):
            env_config = yaml.load(open(filename))
            break
    else:
        env_config = {}
        filename = os.path.abspath('.')
    msg("loading env config from {!r}".format(os.path.abspath(filename)))

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
        return {
            'shell': SUPPORTED_SHELLS[0],
            'modifiers': [],
        }


def main(argv):
    try:
        parser = argparse.ArgumentParser(description='Set up shell environments for projects')
        parser.add_argument('--cfd', type=int)
        parser.add_argument('path')
        parser.add_argument('--env-config')

        args = parser.parse_args(argv[1:])

        config = get_config(args)

        context = load_context(args)
        for ctx_row in config['context']:
            update_context(context, ctx_row)
        
        if args.path == '':
            do_shell(config, context)
        else:
            if '/' in args.path:
                elt, path = args.path.split('/', 1)
            else:
                elt, path = args.path, ''
            start_child(config, context, elt, path)
    except SystemExit:
        pass

    # if we end up here, something went wrong, so hang around long enough to see the message
    raw_input("ENTER to exit ")


if __name__ == "__main__":
    main(sys.argv)
