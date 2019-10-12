<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateUsersGroupsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('users_groups', function (Blueprint $table) {
            $table->unsignedInteger('user_id');
            $table->string('group_key', 10);

            $table->unique(['user_id', 'group_key']);

            $table->index('user_id');
            $table->index('group_key');

            $table->foreign('user_id')->references('id')->on('users');
            $table->foreign('group_key')->references('group_key')->on('groups');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('users_groups');
    }
}
